// user_repo.go 实现 domain.UserRepository 和 domain.TokenBlacklistRepository。
//
// UserRepo        — 用户 CRUD，基于 MySQL/Postgres（GORM）
// TokenBlacklistRepo — Token 登出黑名单，基于 Redis（key 带 TTL）
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/internal/domain"
	pkgredis "github.com/jiujuan/go-ants/pkg/redis"
)

// =============================================================================
// UserRepo — 用户仓储
// =============================================================================

// UserRepo 实现 domain.UserRepository，封装 users 表的读写操作。
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建 UserRepo，从 Data 容器中获取 DB 连接。
func NewUserRepo(data *Data) domain.UserRepository {
	return &UserRepo{db: data.db.GetDB()}
}

// Create 插入新用户记录。
func (r *UserRepo) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// GetByID 根据主键查询用户，不存在时返回 domain.ErrUserNotFound。
func (r *UserRepo) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱查询用户，不存在时返回 domain.ErrUserNotFound。
func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名查询用户，不存在时返回 domain.ErrUserNotFound。
func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// ExistsByEmail 判断邮箱是否已被注册。
func (r *UserRepo) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询邮箱失败: %w", err)
	}
	return count > 0, nil
}

// ExistsByUsername 判断用户名是否已被注册。
func (r *UserRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询用户名失败: %w", err)
	}
	return count > 0, nil
}

// =============================================================================
// TokenBlacklistRepo — Token 登出黑名单（基于 Redis）
// =============================================================================

// tokenBlacklistPrefix 是 Redis key 的命名空间前缀。
// 完整 key 格式：token:blacklist:<access_token>
const tokenBlacklistPrefix = "token:blacklist:"

// TokenBlacklistRepo 实现 domain.TokenBlacklistRepository。
// 登出时将 Access Token 写入 Redis，TTL 等于 Token 剩余有效期；
// 鉴权时检查该 key 是否存在，存在则拒绝请求。
type TokenBlacklistRepo struct {
	redis *pkgredis.Client
}

// NewTokenBlacklistRepo 创建 TokenBlacklistRepo，从 Data 容器中获取 Redis 客户端。
func NewTokenBlacklistRepo(data *Data) domain.TokenBlacklistRepository {
	return &TokenBlacklistRepo{redis: data.redis}
}

// Add 将 token 写入黑名单，ttlSeconds 为剩余有效秒数（到期自动删除）。
func (r *TokenBlacklistRepo) Add(token string, ttlSeconds int64) error {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	expiration := time.Duration(ttlSeconds) * time.Second
	if err := r.redis.Set(ctx, key, "1", expiration); err != nil {
		return fmt.Errorf("加入黑名单失败: %w", err)
	}
	return nil
}

// Exists 判断 token 是否在黑名单中。
// Redis 连接异常时返回 false（降级放行），避免因缓存故障阻断正常请求。
func (r *TokenBlacklistRepo) Exists(token string) (bool, error) {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	n, err := r.redis.Exists(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("查询黑名单失败: %w", err)
	}
	return n > 0, nil
}
