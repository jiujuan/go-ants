package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/jiujuan/go-ants/pkg/log"
)

// DB 数据库实例包装器
type DB struct {
	*gorm.DB
	driver DBType
}

// GetDBType 获取数据库类型
func (d *DB) GetDBType() DBType {
	return d.driver
}

// Option 是数据库选项函数
type Option func(*Options)

// Options 数据库配置选项
type Options struct {
	// DSN 数据源名称
	DSN string
	// Driver 驱动类型: mysql, postgres
	Driver DBType
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int
	// MaxOpenConns 最大打开连接数
	MaxOpenConns int
	// ConnMaxLifetime 连接最大生命周期
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大空闲时间
	ConnMaxIdleTime time.Duration
	// SlowThreshold 慢查询阈值
	SlowThreshold time.Duration
	// LogLevel 日志级别
	LogLevel logger.LogLevel
	// NamingStrategy 命名策略
	NamingStrategy schema.NamingStrategy
	// TablePrefix 表前缀
	TablePrefix string
	// DisableForeignKeyConstraint 是否禁用外键约束
	DisableForeignKeyConstraint bool
	// SkipDefaultTransaction 是否跳过默认事务
	SkipDefaultTransaction bool
	// PrepareStmt 是否预编译语句
	PrepareStmt bool
	// Debug 是否开启调试模式
	Debug bool
	// WithContext 是否启用 context 支持
	WithContext bool
	// Callbacks 回调函数
	Callbacks []CallBack
}

// CallBack 回调函数类型
type CallBack func(*gorm.DB)

// WithDSN 设置数据源名称
func WithDSN(dsn string) Option {
	return func(o *Options) {
		o.DSN = dsn
	}
}

// WithDriver 设置驱动类型
func WithDriver(driver DBType) Option {
	return func(o *Options) {
		o.Driver = driver
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(o *Options) {
		o.MaxIdleConns = maxIdleConns
	}
}

// WithMaxOpenConns 设置最大打开连接数
func WithMaxOpenConns(maxOpenConns int) Option {
	return func(o *Options) {
		o.MaxOpenConns = maxOpenConns
	}
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(lifetime time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxLifetime = lifetime
	}
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(idleTime time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxIdleTime = idleTime
	}
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold(threshold time.Duration) Option {
	return func(o *Options) {
		o.SlowThreshold = threshold
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level logger.LogLevel) Option {
	return func(o *Options) {
		o.LogLevel = level
	}
}

// WithNamingStrategy 设置命名策略
func WithNamingStrategy(strategy schema.NamingStrategy) Option {
	return func(o *Options) {
		o.NamingStrategy = strategy
	}
}

// WithTablePrefix 设置表前缀
func WithTablePrefix(prefix string) Option {
	return func(o *Options) {
		o.TablePrefix = prefix
	}
}

// WithDisableForeignKeyConstraint 禁用外键约束
func WithDisableForeignKeyConstraint(disable bool) Option {
	return func(o *Options) {
		o.DisableForeignKeyConstraint = disable
	}
}

// WithSkipDefaultTransaction 跳过默认事务
func WithSkipDefaultTransaction(skip bool) Option {
	return func(o *Options) {
		o.SkipDefaultTransaction = skip
	}
}

// WithPrepareStmt 预编译语句
func WithPrepareStmt(prepare bool) Option {
	return func(o *Options) {
		o.PrepareStmt = prepare
	}
}

// WithDebug 开启调试模式
func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.Debug = debug
	}
}

// WithContext 启用 context 支持
func WithContext(enable bool) Option {
	return func(o *Options) {
		o.WithContext = enable
	}
}

// WithCallback 添加回调函数
func WithCallback(callback CallBack) Option {
	return func(o *Options) {
		o.Callbacks = append(o.Callbacks, callback)
	}
}

// New 创建新的数据库实例
func New(ctx context.Context, opts ...Option) (*DB, error) {
	options := &Options{
		Driver:          DBTypeMySQL,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
		SlowThreshold:   time.Second,
		LogLevel:        logger.Warn,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
		DisableForeignKeyConstraint: false,
		SkipDefaultTransaction:      false,
		PrepareStmt:                 true,
		Debug:                       false,
		WithContext:                 true,
	}

	for _, opt := range opts {
		opt(options)
	}

	// 创建 GORM 配置
	config := &gorm.Config{
		NamingStrategy:              options.NamingStrategy,
		FullSaveAssociations:        false,
		SkipDefaultTransaction:      options.SkipDefaultTransaction,
		PrepareStmt:                 options.PrepareStmt,
		DisableForeignKeyConstraint: options.DisableForeignKeyConstraint,
	}

	// 设置日志
	if options.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		config.Logger = logger.Default.LogMode(options.LogLevel)
	}

	// 创建数据库连接
	var db *gorm.DB
	var err error

	switch options.Driver {
	case DBTypeMySQL:
		db, err = gorm.Open(mysql.Open(options.DSN), config)
	case DBTypePostgres:
		db, err = gorm.Open(postgres.Open(options.DSN), config)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", options.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层 SQL 数据库
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(options.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(options.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(options.ConnMaxIdleTime)

	// 执行回调函数
	for _, callback := range options.Callbacks {
		callback(db)
	}

	// 启用 context 支持
	if options.WithContext {
		db = db.Session(&gorm.Session{
			Context: ctx,
		})
	}

	log.Info("database connected",
		log.String("driver", string(options.Driver)),
		log.String("dsn", maskDSN(options.DSN)))

	return &DB{DB: db, driver: options.Driver}, nil
}

// AutoMigrate 自动迁移数据库表
func (d *DB) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}

// Transaction 执行事务
func (d *DB) Transaction(fn func(*DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return fn(&DB{DB: tx, driver: d.driver})
	})
}

// WithContext 设置 context
func (d *DB) WithContext(ctx context.Context) *DB {
	return &DB{d.DB.WithContext(ctx), driver: d.driver}
}

// GetDB 获取底层 *gorm.DB 实例
func (d *DB) GetDB() *gorm.DB {
	return d.DB
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// maskDSN 辅助函数：隐藏 DSN 中的密码
func maskDSN(dsn string) string {
	return dsn
}

// 保留旧的类型别名以确保向后兼容
type (
	// MySQLOptions MySQL 选项（已移至 mysql.go）
	MySQLOptions = MySQLOptions
	// MySQLOption MySQL 选项函数（已移至 mysql.go）
	MySQLOption = MySQLOption
	// PostgreSQLOptions PostgreSQL 选项（已移至 postgres.go）
	PostgreSQLOptions = PostgreSQLOptions
	// PostgreSQLOption PostgreSQL 选项函数（已移至 postgres.go）
	PostgreSQLOption = PostgreSQLOption
)
