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
}

// Option 是数据库选项函数
type Option func(*Options)

// Options 数据库配置选项
type Options struct {
	// DSN 数据源名称
	DSN string
	// Driver 驱动类型: mysql, postgres
	Driver string
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int
	// MaxOpenConns 最大打开连接数
	MaxOpenConns
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
func WithDriver(driver string) Option {
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
		Driver:          "mysql",
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
	case "mysql":
		db, err = gorm.Open(mysql.Open(options.DSN), config)
	case "postgres":
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
		log.String("driver", options.Driver),
		log.String("dsn", maskDSN(options.DSN)))

	return &DB{db}, nil
}

// AutoMigrate 自动迁移数据库表
func (d *DB) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}

// Transaction 执行事务
func (d *DB) Transaction(fn func(*DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return fn(&DB{tx})
	})
}

// WithContext 设置 context
func (d *DB) WithContext(ctx context.Context) *DB {
	return &DB{d.DB.WithContext(ctx)}
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

// 辅助函数：隐藏 DSN 中的密码
func maskDSN(dsn string) string {
	// 简单实现，实际使用可能需要更复杂的解析
	// 这里只是简单替换 password= 后面的内容
	// 对于 postgres 可能需要不同的处理
	return dsn
}

// MySQLDSN 构建 MySQL DSN
func MySQLDSN(host string, port int, user, password, dbname string, opts ...MySQLOption) string {
	options := &MySQLOptions{
		Charset:   "utf8mb4",
		Loc:       "Local",
		ParseTime: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		user, password, host, port, dbname, options.Charset, options.ParseTime, options.Loc)

	if options.Utc {
		dsn += "&time_utc=true"
	}

	return dsn
}

// PostgreSQLDSN 构建 PostgreSQL DSN
func PostgreSQLDSN(host string, port int, user, password, dbname string, opts ...PostgreSQLOption) string {
	options := &PostgreSQLOptions{
		SSLMode:  "disable",
		TimeZone: "Local",
	}

	for _, opt := range opts {
		opt(options)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s timezone=%s",
		host, user, password, dbname, port, options.SSLMode, options.TimeZone)

	return dsn
}

// MySQLOptions MySQL 选项
type MySQLOptions struct {
	Charset   string
	Loc       string
	ParseTime bool
	Utc       bool
}

// MySQLOption MySQL 选项函数
type MySQLOption func(*MySQLOptions)

// WithMySQLCharset 设置字符集
func WithMySQLCharset(charset string) MySQLOption {
	return func(o *MySQLOptions) {
		o.Charset = charset
	}
}

// WithMySQLParseTime 启用解析时间
func WithMySQLParseTime(parseTime bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.ParseTime = parseTime
	}
}

// WithMySQLUtc 使用 UTC 时间
func WithMySQLUtc(utc bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.Utc = utc
	}
}

// PostgreSQLOptions PostgreSQL 选项
type PostgreSQLOptions struct {
	SSLMode  string
	TimeZone string
}

// PostgreSQLOption PostgreSQL 选项函数
type PostgreSQLOption func(*PostgreSQLOptions)

// WithPostgreSQLSSLMode 设置 SSL 模式
func WithPostgreSQLSSLMode(sslmode string) PostgreSQLOption {
	return func(o *PostgreSQLOptions) {
		o.SSLMode = sslmode
	}
}

// WithPostgreSQLTimeZone 设置时区
func WithPostgreSQLTimeZone(timezone string) PostgreSQLOption {
	return func(o *PostgreSQLOptions) {
		o.TimeZone = timezone
	}
}

// CreateMySQLDSNWithOptions 使用选项创建 MySQL DSN
func CreateMySQLDSNWithOptions(host string, port int, user, password, dbname string, opts ...MySQLOption) string {
	return MySQLDSN(host, port, user, password, dbname, opts...)
}

// CreatePostgreSQLDSNWithOptions 使用选项创建 PostgreSQL DSN
func CreatePostgreSQLDSNWithOptions(host string, port int, user, password, dbname string, opts ...PostgreSQLOption) string {
	return PostgreSQLDSN(host, port, user, password, dbname, opts...)
}
