// Package data 提供了数据访问层，负责与数据库、缓存等数据存储进行交互。
package data

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/pkg/database"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/redis"
)

// Data 数据层实例
type Data struct {
	db    *database.DB
	redis *redis.Client
	log   *log.Logger
}

// New 创建数据层实例
func New(db *database.DB, redisClient *redis.Client) (*Data, func(), error) {
	d := &Data{
		db:    db,
		redis: redisClient,
		log:   log.DefaultLogger(),
	}

	cleanup := func() {
		if d.redis != nil {
			d.redis.Close()
		}
		if d.db != nil {
			d.db.Close()
		}
	}

	return d, cleanup, nil
}

// DB 获取数据库实例
func (d *Data) DB() *database.DB {
	return d.db
}

// Redis 获取 Redis 实例
func (d *Data) Redis() *redis.Client {
	return d.redis
}

// WithContext 使用 context
func (d *Data) WithContext(ctx context.Context) *Data {
	return &Data{
		db:    d.db.WithContext(ctx),
		redis: d.redis,
		log:   d.log,
	}
}

// ===== 数据仓库实现示例 =====

// Repo 数据仓库接口
type Repo interface {
	Create(ctx context.Context, entity interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}) error
	GetByID(ctx context.Context, id interface{}) (interface{}, error)
	List(ctx context.Context, opts ...ListOption) ([]interface{}, error)
}

// ListOption 列表查询选项
type ListOption func(*ListOptions)

// ListOptions 列表查询配置
type ListOptions struct {
	Page     int
	PageSize int
	OrderBy  string
	Filters  map[string]interface{}
}

// BaseRepo 基础仓库
type BaseRepo struct {
	db *gorm.DB
}

// NewBaseRepo 创建基础仓库
func NewBaseRepo(db *database.DB) *BaseRepo {
	return &BaseRepo{
		db: db.GetDB(),
	}
}
