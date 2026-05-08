// Package service 提供应用服务层，编排用户注册、登录、登出业务逻辑。
package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/auth"
	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== DTO 定义 =====

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Nickname string `json:"nickname" binding:"max=100"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求（支持用户名或邮箱登录）
type LoginRequest struct {
	// Account 可填用户名或邮箱
	Account  string `json:"account"  binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"` // Unix 时间戳
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	// AccessToken 从 Authorization header 中提取，由 middleware 注入
	AccessToken string `json:"-"`
}

// ===== 服务接口 =====

// UserService 用户服务接口
type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) error
}

// ===== 服务实现 =====

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	userRepo      domain.UserRepository
	blacklistRepo domain.TokenBlacklistRepository
	userDomain    *domain.UserDomain
	jwtAuth       *auth.JWT
	log           *log.Logger
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo domain.UserRepository,
	blacklistRepo domain.TokenBlacklistRepository,
	jwtAuth *auth.JWT,
) UserService {
	return &UserServiceImpl{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		userDomain:    domain.NewUserDomain(),
		jwtAuth:       jwtAuth,
		log:           log.DefaultLogger(),
	}
}

// Register 用户注册
func (s *UserServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. 领域规则校验
	if err := s.userDomain.ValidateUsername(req.Username); err != nil {
		return nil, &domain.DomainError{Code: domain.ErrCodeInvalidUsername, Message: err.Error()}
	}
	if err := s.userDomain.ValidatePassword(req.Password); err != nil {
		return nil, &domain.DomainError{Code: domain.ErrCodeInvalidPassword, Message: err.Error()}
	}

	// 2. 唯一性检查
	if exists, err := s.userRepo.ExistsByUsername(req.Username); err != nil {
		return nil, fmt.Errorf("服务异常，请稍后重试")
	} else if exists {
		return nil, domain.ErrUserAlreadyExist
	}

	if exists, err := s.userRepo.ExistsByEmail(req.Email); err != nil {
		return nil, fmt.Errorf("服务异常，请稍后重试")
	} else if exists {
		return nil, &domain.DomainError{Code: domain.ErrCodeUserAlreadyExist, Message: "该邮箱已被注册"}
	}

	// 3. 密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Errorw("hash password failed", "err", err)
		return nil, fmt.Errorf("服务异常，请稍后重试")
	}

	// 4. 持久化
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		Nickname:     nickname,
		Status:       domain.UserStatusActive,
	}
	if err := s.userRepo.Create(user); err != nil {
		s.log.Errorw("create user failed", "username", req.Username, "err", err)
		return nil, fmt.Errorf("注册失败，请稍后重试")
	}

	s.log.Infow("user registered", "userID", user.ID, "username", user.Username)
	return &RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
	}, nil
}

// Login 用户登录（支持用户名或邮箱）
func (s *UserServiceImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 1. 查找用户（先尝试邮箱，再尝试用户名）
	var user *domain.User
	var err error

	// 判断是否含 @ 符号，有则当邮箱处理
	if isEmail(req.Account) {
		user, err = s.userRepo.GetByEmail(req.Account)
	} else {
		user, err = s.userRepo.GetByUsername(req.Account)
	}
	if err != nil {
		if isDomainError(err, domain.ErrCodeUserNotFound) {
			// 用户不存在，返回通用错误，避免枚举
			return nil, &domain.DomainError{Code: domain.ErrCodeInvalidPassword, Message: "账号或密码不正确"}
		}
		s.log.Errorw("get user failed", "account", req.Account, "err", err)
		return nil, fmt.Errorf("服务异常，请稍后重试")
	}

	// 2. 检查账户状态
	if user.Status == domain.UserStatusInactive {
		return nil, domain.ErrUserDisabled
	}

	// 3. 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &domain.DomainError{Code: domain.ErrCodeInvalidPassword, Message: "账号或密码不正确"}
	}

	// 4. 生成 JWT Token
	userInfo := &auth.UserInfo{
		ID:       strconv.FormatUint(uint64(user.ID), 10),
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}
	tokenInfo, err := s.jwtAuth.GenerateToken(ctx, userInfo, 0) // 0 使用默认过期时间
	if err != nil {
		s.log.Errorw("generate token failed", "userID", user.ID, "err", err)
		return nil, fmt.Errorf("服务异常，请稍后重试")
	}

	s.log.Infow("user logged in", "userID", user.ID, "username", user.Username)
	return &LoginResponse{
		AccessToken:  tokenInfo.AccessToken,
		RefreshToken: tokenInfo.RefreshToken,
		TokenType:    tokenInfo.TokenType,
		ExpiresAt:    tokenInfo.ExpiresAt.Unix(),
		UserID:       user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
	}, nil
}

// Logout 用户登出（将 Token 加入黑名单）
func (s *UserServiceImpl) Logout(ctx context.Context, req *LogoutRequest) error {
	if req.AccessToken == "" {
		return nil // token 为空，幂等处理
	}

	// 解析 token 获取剩余有效期
	claims, err := s.jwtAuth.VerifyToken(ctx, req.AccessToken)
	if err != nil {
		// token 已过期或无效，无需加黑名单，直接返回成功
		return nil
	}

	// 计算剩余 TTL（秒）
	ttl := int64(time.Until(claims.ExpiresAt.Time).Seconds())
	if ttl <= 0 {
		return nil // 已过期，无需操作
	}

	if err := s.blacklistRepo.Add(req.AccessToken, ttl); err != nil {
		s.log.Errorw("add token to blacklist failed", "err", err)
		// 黑名单写入失败不阻断登出流程，仅记录日志
	}

	s.log.Infow("user logged out", "subject", claims.Subject)
	return nil
}

// ===== 工具函数 =====

func isEmail(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

func isDomainError(err error, code domain.ErrorCode) bool {
	if de, ok := err.(*domain.DomainError); ok {
		return de.Code == code
	}
	return false
}
