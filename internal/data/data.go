// Package data 提供数据访问层，实现 domain 层定义的仓储接口，负责与 DB/Redis 交互。
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/database"
	pkgredis "github.com/jiujuan/go-ants/pkg/redis"
	"github.com/jiujuan/go-ants/pkg/log"
)

// Data 数据层容器，持有 DB 和 Redis 客户端
type Data struct {
	db    *database.DB
	redis *pkgredis.Client
	log   *log.Logger
}

// New 创建数据层实例
func New(db *database.DB, redisClient *pkgredis.Client) (*Data, func(), error) {
	d := &Data{
		db:    db,
		redis: redisClient,
		log:   log.DefaultLogger(),
	}
	cleanup := func() {
		if d.db != nil {
			if err := d.db.Close(); err != nil {
				d.log.Errorw("close db failed", "err", err)
			}
		}
		if d.redis != nil {
			if err := d.redis.Close(); err != nil {
				d.log.Errorw("close redis failed", "err", err)
			}
		}
	}
	return d, cleanup, nil
}

// ===== 用户仓储实现 =====

// UserRepo 用户数据仓储
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储
func NewUserRepo(data *Data) domain.UserRepository {
	return &UserRepo{db: data.db.GetDB()}
}

func (r *UserRepo) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

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

func (r *UserRepo) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询邮箱失败: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询用户名失败: %w", err)
	}
	return count > 0, nil
}

// ===== Token 黑名单仓储实现（基于 Redis）=====

const tokenBlacklistPrefix = "token:blacklist:"

// TokenBlacklistRepo 基于 Redis 的 Token 黑名单仓储
type TokenBlacklistRepo struct {
	redis *pkgredis.Client
}

// NewTokenBlacklistRepo 创建 Token 黑名单仓储
func NewTokenBlacklistRepo(data *Data) domain.TokenBlacklistRepository {
	return &TokenBlacklistRepo{redis: data.redis}
}

// Add 将 token 加入黑名单，ttl 为剩余有效秒数
func (r *TokenBlacklistRepo) Add(token string, ttlSeconds int64) error {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	expiration := time.Duration(ttlSeconds) * time.Second
	if err := r.redis.Set(ctx, key, "1", expiration); err != nil {
		return fmt.Errorf("加入黑名单失败: %w", err)
	}
	return nil
}

// Exists 判断 token 是否在黑名单中
func (r *TokenBlacklistRepo) Exists(token string) (bool, error) {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	n, err := r.redis.Exists(ctx, key)
	if err != nil {
		// Redis 不可用时，为安全起见返回 false（不阻断正常登录）
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("查询黑名单失败: %w", err)
	}
	return n > 0, nil
}
