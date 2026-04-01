package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ===== PostgreSQL 驱动 =====

// PostgreSQLDriver PostgreSQL 数据库驱动
type PostgreSQLDriver struct{}

// NewPostgreSQLDriver 创建 PostgreSQL 驱动
func NewPostgreSQLDriver() *PostgreSQLDriver {
	return &PostgreSQLDriver{}
}

// GetDBType 获取数据库类型
func (d *PostgreSQLDriver) GetDBType() DBType {
	return DBTypePostgres
}

// Open 打开数据库连接
func (d *PostgreSQLDriver) Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// ===== PostgreSQL DSN 构建 =====

// PostgreSQLOptions PostgreSQL 选项
type PostgreSQLOptions struct {
	SSLMode     string
	TimeZone    string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
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

// WithPostgreSQLSSLCert 设置 SSL 证书路径
func WithPostgreSQLSSLCert(cert string) PostgreSQLOption {
	return func(o *PostgreSQLOptions) {
		o.SSLCert = cert
	}
}

// WithPostgreSQLSSLKey 设置 SSL 密钥路径
func WithPostgreSQLSSLKey(key string) PostgreSQLOption {
	return func(o *PostgreSQLOptions) {
		o.SSLKey = key
	}
}

// WithPostgreSQLSSLRootCert 设置 SSL 根证书路径
func WithPostgreSQLSSLRootCert(rootCert string) PostgreSQLOption {
	return func(o *PostgreSQLOptions) {
		o.SSLRootCert = rootCert
	}
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

	if options.SSLCert != "" {
		dsn += fmt.Sprintf(" sslcert=%s", options.SSLCert)
	}
	if options.SSLKey != "" {
		dsn += fmt.Sprintf(" sslkey=%s", options.SSLKey)
	}
	if options.SSLRootCert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", options.SSLRootCert)
	}

	return dsn
}

// CreatePostgreSQLDSNWithOptions 使用选项创建 PostgreSQL DSN
func CreatePostgreSQLDSNWithOptions(host string, port int, user, password, dbname string, opts ...PostgreSQLOption) string {
	return PostgreSQLDSN(host, port, user, password, dbname, opts...)
}

// ===== PostgreSQL 连接配置 =====

// PostgreSQLConfig PostgreSQL 连接配置
type PostgreSQLConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	TimeZone    string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

// NewPostgreSQLConfig 创建 PostgreSQL 配置
func NewPostgreSQLConfig(host string, port int, user, password, database string, opts ...PostgreSQLOption) *PostgreSQLConfig {
	config := &PostgreSQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  "disable",
		TimeZone: "Local",
	}

	for _, opt := range opts {
		opt((*PostgreSQLOptions)(nil))
	}

	return config
}

// GetDSN 获取 DSN
func (c *PostgreSQLConfig) GetDSN() string {
	return PostgreSQLDSN(c.Host, c.Port, c.User, c.Password, c.Database,
		WithPostgreSQLSSLMode(c.SSLMode),
		WithPostgreSQLTimeZone(c.TimeZone),
		WithPostgreSQLSSLCert(c.SSLCert),
		WithPostgreSQLSSLKey(c.SSLKey),
		WithPostgreSQLSSLRootCert(c.SSLRootCert),
	)
}

// GetDBType 获取数据库类型
func (c *PostgreSQLConfig) GetDBType() DBType {
	return DBTypePostgres
}

// ===== PostgreSQL 连接器 =====

// PostgreSQLConnector PostgreSQL 连接器
type PostgreSQLConnector struct {
	config *PostgreSQLConfig
}

// NewPostgreSQLConnector 创建 PostgreSQL 连接器
func NewPostgreSQLConnector(config *PostgreSQLConfig) *PostgreSQLConnector {
	return &PostgreSQLConnector{config: config}
}

// Connect 创建数据库连接
func (c *PostgreSQLConnector) Connect() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(c.config.GetDSN()), &gorm.Config{})
}

// ===== PostgreSQL 工厂 =====

// PostgreSQLFactory PostgreSQL 工厂
type PostgreSQLFactory struct{}

// NewPostgreSQLFactory 创建 PostgreSQL 工厂
func NewPostgreSQLFactory() *PostgreSQLFactory {
	return &PostgreSQLFactory{}
}

// Create 创建 PostgreSQL 连接
func (f *PostgreSQLFactory) Create(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// CreateWithConfig 使用配置创建 PostgreSQL 连接
func (f *PostgreSQLFactory) CreateWithConfig(config *PostgreSQLConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(config.GetDSN()), &gorm.Config{})
}

// ===== PostgreSQL 辅助函数 =====

// NewPostgreSQL 创建 PostgreSQL 数据库实例的便捷函数
func NewPostgreSQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// IsPostgreSQLDSN 检查是否为 PostgreSQL DSN
func IsPostgreSQLDSN(dsn string) bool {
	return len(dsn) > 0 && (contains(dsn, "postgres") || contains(dsn, "host="))
}

// ===== 辅助函数 =====

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

// containsHelper 辅助函数
func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
