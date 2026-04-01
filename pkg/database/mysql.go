package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ===== MySQL 驱动 =====

// MySQLDriver MySQL 数据库驱动
type MySQLDriver struct{}

// NewMySQLDriver 创建 MySQL 驱动
func NewMySQLDriver() *MySQLDriver {
	return &MySQLDriver{}
}

// GetDBType 获取数据库类型
func (d *MySQLDriver) GetDBType() DBType {
	return DBTypeMySQL
}

// Open 打开数据库连接
func (d *MySQLDriver) Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// ===== MySQL DSN 构建 =====

// MySQLOptions MySQL 选项
type MySQLOptions struct {
	Charset   string
	Loc       string
	ParseTime bool
	Utc       bool
	Timeout   string
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

// WithMySQLTimeout 设置超时时间
func WithMySQLTimeout(timeout string) MySQLOption {
	return func(o *MySQLOptions) {
		o.Timeout = timeout
	}
}

// WithMySQLLoc 设置时区
func WithMySQLLoc(loc string) MySQLOption {
	return func(o *MySQLOptions) {
		o.Loc = loc
	}
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

	if options.Timeout != "" {
		dsn += "&timeout=" + options.Timeout
	}

	return dsn
}

// CreateMySQLDSNWithOptions 使用选项创建 MySQL DSN
func CreateMySQLDSNWithOptions(host string, port int, user, password, dbname string, opts ...MySQLOption) string {
	return MySQLDSN(host, port, user, password, dbname, opts...)
}

// ===== MySQL 连接配置 =====

// MySQLConfig MySQL 连接配置
type MySQLConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	Charset   string
	ParseTime bool
	Loc       string
	Timeout   string
}

// NewMySQLConfig 创建 MySQL 配置
func NewMySQLConfig(host string, port int, user, password, database string, opts ...MySQLOption) *MySQLConfig {
	config := &MySQLConfig{
		Host:      host,
		Port:      port,
		User:      user,
		Password:  password,
		Database:  database,
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	for _, opt := range opts {
		switch opt {
		}
	}

	return config
}

// GetDSN 获取 DSN
func (c *MySQLConfig) GetDSN() string {
	return MySQLDSN(c.Host, c.Port, c.User, c.Password, c.Database,
		WithMySQLCharset(c.Charset),
		WithMySQLParseTime(c.ParseTime),
		WithMySQLLoc(c.Loc),
		WithMySQLTimeout(c.Timeout),
	)
}

// GetDBType 获取数据库类型
func (c *MySQLConfig) GetDBType() DBType {
	return DBTypeMySQL
}

// ===== MySQL 连接器 =====

// MySQLConnector MySQL 连接器
type MySQLConnector struct {
	config *MySQLConfig
}

// NewMySQLConnector 创建 MySQL 连接器
func NewMySQLConnector(config *MySQLConfig) *MySQLConnector {
	return &MySQLConnector{config: config}
}

// Connect 创建数据库连接
func (c *MySQLConnector) Connect() (*gorm.DB, error) {
	return gorm.Open(mysql.Open(c.config.GetDSN()), &gorm.Config{})
}

// ===== MySQL 工厂 =====

// MySQLFactory MySQL 工厂
type MySQLFactory struct{}

// NewMySQLFactory 创建 MySQL 工厂
func NewMySQLFactory() *MySQLFactory {
	return &MySQLFactory{}
}

// Create 创建 MySQL 连接
func (f *MySQLFactory) Create(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// CreateWithConfig 使用配置创建 MySQL 连接
func (f *MySQLFactory) CreateWithConfig(config *MySQLConfig) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(config.GetDSN()), &gorm.Config{})
}

// ===== MySQL 辅助函数 =====

// NewMySQL 创建 MySQL 数据库实例的便捷函数
func NewMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// IsMySQLDSN 检查是否为 MySQL DSN
func IsMySQLDSN(dsn string) bool {
	return len(dsn) > 0 && (contains(dsn, "mysql") || contains(dsn, "@tcp"))
}
