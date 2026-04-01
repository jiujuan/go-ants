package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ===== 数据库类型 =====

// DBType 数据库类型枚举
type DBType string

const (
	DBTypeMySQL     DBType = "mysql"
	DBTypePostgres  DBType = "postgres"
	DBTypeSQLite    DBType = "sqlite"
	DBTypeSQLServer DBType = "sqlserver"
)

// ===== 数据库驱动接口 =====

// Driver 数据库驱动接口
type Driver interface {
	// GetDBType 获取数据库类型
	GetDBType() DBType
	// Open 打开数据库连接
	Open(dsn string) (*gorm.DB, error)
}

// DriverFactory 驱动工厂函数类型
type DriverFactory func() Driver

// ===== 数据库接口 =====

// Database 数据库操作接口
type Database interface {
	// GetDB 获取底层 *gorm.DB 实例
	GetDB() *gorm.DB
	// Close 关闭数据库连接
	Close() error
	// AutoMigrate 自动迁移数据库表
	AutoMigrate(dest ...interface{}) error
	// Transaction 执行事务
	Transaction(func(*DB) error) error
	// WithContext 设置 context
	WithContext(ctx context.Context) *DB
	// GetDBType 获取数据库类型
	GetDBType() DBType
}

// ===== GORM 回调接口 =====

// BeforeCreateCallback 创建前回调
type BeforeCreateCallback func(db *gorm.DB) error

// AfterCreateCallback 创建后回调
type AfterCreateCallback func(db *gorm.DB) error

// BeforeUpdateCallback 更新前回调
type BeforeUpdateCallback func(db *gorm.DB) error

// AfterUpdateCallback 更新后回调
type AfterUpdateCallback func(db *gorm.DB) error

// BeforeDeleteCallback 删除前回调
type BeforeDeleteCallback func(db *gorm.DB) error

// AfterDeleteCallback 删除后回调
type AfterDeleteCallback func(db *gorm.DB) error

// AfterFindCallback 查询后回调
type AfterFindCallback func(db *gorm.DB) error

// ===== 查询构建器接口 =====

// QueryBuilder 查询构建器接口
type QueryBuilder interface {
	// Where 添加 WHERE 条件
	Where(query interface{}, args ...interface{}) QueryBuilder
	// Order 添加 ORDER BY
	Order(value interface{}) QueryBuilder
	// Limit 添加 LIMIT
	Limit(limit int) QueryBuilder
	// Offset 添加 OFFSET
	Offset(offset int) QueryBuilder
	// Find 查询多条记录
	Find(dest interface{}) error
	// First 查询单条记录
	First(dest interface{}) error
	// Create 创建记录
	Create(dest interface{}) error
	// Updates 更新记录
	Updates(values interface{}) error
	// Delete 删除记录
	Delete(dest interface{}) error
	// Count 计数
	Count(count *int64) error
}

// ===== 仓储接口 =====

// Repository 通用仓储接口
type Repository interface {
	// Create 创建
	Create(ctx context.Context, model interface{}) error
	// GetByID 根据ID获取
	GetByID(ctx context.Context, id interface{}, model interface{}) error
	// Update 更新
	Update(ctx context.Context, model interface{}) error
	// Delete 删除
	Delete(ctx context.Context, id interface{}) error
	// List 列表查询
	List(ctx context.Context, query interface{}, page, pageSize int, models interface{}) error
	// Count 计数
	Count(ctx context.Context, query interface{}) (int64, error)
}

// ===== 基础仓储实现 =====

// BaseRepository 基础仓储实现
type BaseRepository struct {
	db *DB
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository(db *DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Create 创建
func (r *BaseRepository) Create(ctx context.Context, model interface{}) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID 根据ID获取
func (r *BaseRepository) GetByID(ctx context.Context, id interface{}, model interface{}) error {
	return r.db.WithContext(ctx).First(model, id).Error
}

// Update 更新
func (r *BaseRepository) Update(ctx context.Context, model interface{}) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete 删除
func (r *BaseRepository) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(id).Error
}

// List 列表查询
func (r *BaseRepository) List(ctx context.Context, query interface{}, page, pageSize int, models interface{}) error {
	offset := (page - 1) * pageSize
	return r.db.WithContext(ctx).Where(query).Offset(offset).Limit(pageSize).Find(models).Error
}

// Count 计数
func (r *BaseRepository) Count(ctx context.Context, query interface{}) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(query).Count(&count).Error
	return count, err
}

// ===== 数据库配置接口 =====

// Config 数据库配置接口
type Config interface {
	// GetDSN 获取数据源名称
	GetDSN() string
	// GetDBType 获取数据库类型
	GetDBType() DBType
	// GetMaxIdleConns 获取最大空闲连接数
	GetMaxIdleConns() int
	// GetMaxOpenConns 获取最大打开连接数
	GetMaxOpenConns() int
	// GetConnMaxLifetime 获取连接最大生命周期
	GetConnMaxLifetime() time.Duration
}

// ===== 通用选项 =====

// CommonOptions 通用数据库选项
type CommonOptions struct {
	DBType          DBType
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	TablePrefix     string
}

// CommonOption 通用选项函数
type CommonOption func(*CommonOptions)

// WithCommonDBType 设置数据库类型
func WithCommonDBType(dbType DBType) CommonOption {
	return func(o *CommonOptions) {
		o.DBType = dbType
	}
}

// WithCommonHost 设置主机
func WithCommonHost(host string) CommonOption {
	return func(o *CommonOptions) {
		o.Host = host
	}
}

// WithCommonPort 设置端口
func WithCommonPort(port int) CommonOption {
	return func(o *CommonOptions) {
		o.Port = port
	}
}

// WithCommonUser 设置用户
func WithCommonUser(user string) CommonOption {
	return func(o *CommonOptions) {
		o.User = user
	}
}

// WithCommonPassword 设置密码
func WithCommonPassword(password string) CommonOption {
	return func(o *CommonOptions) {
		o.Password = password
	}
}

// WithCommonDatabase 设置数据库名
func WithCommonDatabase(database string) CommonOption {
	return func(o *CommonOptions) {
		o.Database = database
	}
}

// WithCommonMaxIdleConns 设置最大空闲连接数
func WithCommonMaxIdleConns(maxIdleConns int) CommonOption {
	return func(o *CommonOptions) {
		o.MaxIdleConns = maxIdleConns
	}
}

// WithCommonMaxOpenConns 设置最大打开连接数
func WithCommonMaxOpenConns(maxOpenConns int) CommonOption {
	return func(o *CommonOptions) {
		o.MaxOpenConns = maxOpenConns
	}
}

// WithCommonConnMaxLifetime 设置连接最大生命周期
func WithCommonConnMaxLifetime(lifetime time.Duration) CommonOption {
	return func(o *CommonOptions) {
		o.ConnMaxLifetime = lifetime
	}
}

// WithCommonTablePrefix 设置表前缀
func WithCommonTablePrefix(prefix string) CommonOption {
	return func(o *CommonOptions) {
		o.TablePrefix = prefix
	}
}

// ===== 错误定义 =====

var (
	ErrDBNotSupported    = fmt.Errorf("database type not supported")
	ErrConnectionFailed  = fmt.Errorf("database connection failed")
	ErrTransactionFailed = fmt.Errorf("transaction failed")
	ErrMigrationFailed   = fmt.Errorf("migration failed")
	ErrQueryFailed       = fmt.Errorf("query failed")
)

// ===== 工具函数 =====

// MaskDSN 隐藏 DSN 中的密码
func MaskDSN(dsn string) string {
	// 这里只是简单返回，实际使用可能需要更复杂的处理
	return dsn
}

// GetSQLDB 获取底层 *sql.DB
func GetSQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// Ping 检查数据库连接
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
