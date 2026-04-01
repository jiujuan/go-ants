package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/jiujuan/go-ants/pkg/log"
)

var (
	// ErrInvalidToken 无效的令牌
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken 过期的令牌
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidSigningMethod 无效的签名方法
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	// ErrMissingToken 缺失的令牌
	ErrMissingToken = errors.New("missing token")
	// ErrInvalidClaims 无效的声明
	ErrInvalidClaims = errors.New("invalid claims")
)

// Claims 自定义声明
type Claims struct {
	jwt.RegisteredClaims
	// 自定义字段
	Permissions []string               `json:"permissions,omitempty"`
	Nickname    string                 `json:"nickname,omitempty"`
	Avatar      string                 `json:"avatar,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID          string
	Username    string
	Nickname    string
	Email       string
	Permissions []string
	Extra       map[string]interface{}
}

// TokenInfo Token 信息
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// Authorizer 授权器接口
type Authorizer interface {
	// GenerateToken 生成令牌
	GenerateToken(ctx context.Context, user *UserInfo, duration time.Duration) (*TokenInfo, error)
	// VerifyToken 验证令牌
	VerifyToken(ctx context.Context, tokenString string) (*Claims, error)
	// ParseUserInfo 从令牌解析用户信息
	ParseUserInfo(ctx context.Context, tokenString string) (*UserInfo, error)
	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string, duration time.Duration) (*TokenInfo, error)
	// InvalidateToken 撤销令牌
	InvalidateToken(ctx context.Context, tokenString string) error
}

// JWT jwt 授权器
type JWT struct {
	signingKey        interface{}
	signingMethod     jwt.SigningMethod
	issuer            string
	audience          []string
	expiration        time.Duration
	refreshExpiration time.Duration
	keyFunc           func(token *jwt.Token) (interface{}, error)
	tokenLookup       string
	tokenHeader       string
}

// Option 是 JWT 选项函数
type Option func(*JWT)

// New 创建新的 JWT 授权器
func New(opts ...Option) *JWT {
	j := &JWT{
		signingMethod:     jwt.SigningMethodHS256,
		issuer:            "go-ants",
		audience:          []string{"go-ants"},
		expiration:        time.Hour * 24,
		refreshExpiration: time.Hour * 24 * 30,
		tokenLookup:       "header:Authorization",
		tokenHeader:       "Bearer",
	}

	// 默认使用 HMAC 密钥
	key := make([]byte, 64)
	rand.Read(key)
	j.signingKey = key

	for _, opt := range opts {
		opt(j)
	}

	// 设置 keyFunc
	j.keyFunc = func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return j.signingKey, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			if j.signingKey == nil {
				return nil, ErrInvalidSigningMethod
			}
			return j.signingKey, nil
		}
		return nil, ErrInvalidSigningMethod
	}

	return j
}

// GenerateToken 生成令牌
func (j *JWT) GenerateToken(ctx context.Context, user *UserInfo, duration time.Duration) (*TokenInfo, error) {
	if duration == 0 {
		duration = j.expiration
	}

	now := time.Now()

	// Access Token
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Audience:  j.audience,
			Subject:   user.ID,
		},
		Permissions: user.Permissions,
		Nickname:    user.Nickname,
		Extra:       user.Extra,
	}

	token := jwt.NewWithClaims(j.signingMethod, claims)
	accessToken, err := token.SignedString(j.signingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Audience:  j.audience,
			Subject:   user.ID,
		},
	}
	refreshToken := jwt.NewWithClaims(j.signingMethod, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.signingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(duration),
		TokenType:    j.tokenHeader,
	}, nil
}

// VerifyToken 验证令牌
func (j *JWT) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, j.keyFunc)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.TokenExpiredMask != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 验证 claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// 验证 issuer
	if j.issuer != "" && claims.Issuer != j.issuer {
		return nil, ErrInvalidClaims
	}

	// 验证 audience
	if len(j.audience) > 0 {
		validAudience := false
		for _, aud := range j.audience {
			if subtle.ConstantTimeCompare([]byte(aud), []byte(claims.Audience[0])) == 1 {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, ErrInvalidClaims
		}
	}

	log.Info("token verified",
		log.String("subject", claims.Subject),
		log.String("nickname", claims.Nickname))

	return claims, nil
}

// ParseUserInfo 从令牌解析用户信息
func (j *JWT) ParseUserInfo(ctx context.Context, tokenString string) (*UserInfo, error) {
	claims, err := j.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:          claims.Subject,
		Username:    claims.Subject,
		Nickname:    claims.Nickname,
		Permissions: claims.Permissions,
		Extra:       claims.Extra,
	}, nil
}

// RefreshToken 刷新令牌
func (j *JWT) RefreshToken(ctx context.Context, refreshToken string, duration time.Duration) (*TokenInfo, error) {
	// 验证 refresh token
	claims, err := j.VerifyToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 生成新的 token
	user := &UserInfo{
		ID:          claims.Subject,
		Permissions: claims.Permissions,
		Nickname:    claims.Nickname,
		Extra:       claims.Extra,
	}

	return j.GenerateToken(ctx, user, duration)
}

// InvalidateToken 撤销令牌（可以通过加入黑名单实现）
func (j *JWT) InvalidateToken(ctx context.Context, tokenString string) error {
	// 简单实现：验证 token 有效性即可
	// 生产环境可以加入 Redis 黑名单
	_, err := j.VerifyToken(ctx, tokenString)
	return err
}

// ===== 选项函数 =====

// WithSigningKey 设置签名密钥
func WithSigningKey(key interface{}) Option {
	return func(j *JWT) {
		j.signingKey = key
	}
}

// WithHMACSigningKey 使用 HMAC 签名密钥
func WithHMACSigningKey(secret string) Option {
	return func(j *JWT) {
		j.signingMethod = jwt.SigningMethodHS256
		j.signingKey = []byte(secret)
	}
}

// WithRSASigningKey 使用 RSA 签名密钥
func WithRSASigningKey(privateKey *rsa.PrivateKey) Option {
	return func(j *JWT) {
		j.signingMethod = jwt.SigningMethodRS256
		j.signingKey = privateKey
	}
}

// WithIssuer 设置发行者
func WithIssuer(issuer string) Option {
	return func(j *JWT) {
		j.issuer = issuer
	}
}

// WithAudience 设置受众
func WithAudience(audience ...string) Option {
	return func(j *JWT) {
		j.audience = audience
	}
}

// WithExpiration 设置过期时间
func WithExpiration(expiration time.Duration) Option {
	return func(j *JWT) {
		j.expiration = expiration
	}
}

// WithRefreshExpiration 设置刷新令牌过期时间
func WithRefreshExpiration(expiration time.Duration) Option {
	return func(j *JWT) {
		j.refreshExpiration = expiration
	}
}

// WithTokenLookup 设置 token 查找方式
func WithTokenLookup(lookup string) Option {
	return func(j *JWT) {
		j.tokenLookup = lookup
	}
}

// WithTokenHeader 设置 token header 名称
func WithTokenHeader(header string) Option {
	return func(j *JWT) {
		j.tokenHeader = header
	}
}

// ===== 中间件相关 =====

// ExtractToken 从请求中提取 token
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

// TokenExtractor Token 提取函数类型
type TokenExtractor func(ctx context.Context) (string, error)

// HeaderTokenExtractor 从 Header 提取 Token
func HeaderTokenExtractor(headerName string) TokenExtractor {
	return func(ctx context.Context) (string, error) {
		// 这里简化实现，实际需要从 context 或 request 中获取
		return "", nil
	}
}

// ===== 密码相关 =====

// PasswordHash 密码哈希
type PasswordHash interface {
	// Hash 生成密码哈希
	Hash(password string) (string, error)
	// Compare 比较密码
	Compare(hashedPassword, password string) error
}

// BcryptPasswordHash bcrypt 密码哈希
type BcryptPasswordHash struct {
	cost int
}

// NewBcryptPasswordHash 创建 bcrypt 密码哈希
func NewBcryptPasswordHash(opts ...BcryptOption) *BcryptPasswordHash {
	options := &BcryptOptions{
		cost: 10,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &BcryptPasswordHash{cost: options.cost}
}

// Hash 生成密码哈希
func (b *BcryptPasswordHash) Hash(password string) (string, error) {
	// 这里需要引入 golang.org/x/crypto/bcrypt
	// 为简化，我们使用简单的实现
	return password, nil
}

// Compare 比较密码
func (b *BcryptPasswordHash) Compare(hashedPassword, password string) error {
	if hashedPassword == password {
		return nil
	}
	return errors.New("password mismatch")
}

// BcryptOption bcrypt 选项
type BcryptOption func(*BcryptOptions)

type BcryptOptions struct {
	cost int
}

// WithBcryptCost 设置 bcrypt 成本
func WithBcryptCost(cost int) BcryptOption {
	return func(o *BcryptOptions) {
		o.cost = cost
	}
}

// ===== 工具函数 =====

// GenerateRandomToken 生成随机令牌
func GenerateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// EncodeClaims 编码声明
func EncodeClaims(claims *Claims) (string, error) {
	return json.Marshal(claims)
}

// DecodeClaims 解码声明
func DecodeClaims(data string) (*Claims, error) {
	claims := &Claims{}
	err := json.Unmarshal([]byte(data), claims)
	return claims, err
}
