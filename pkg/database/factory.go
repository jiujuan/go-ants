package database

import (
	"context"
	"fmt"
)

// ===== 数据库工厂管理器 =====

// FactoryManager 数据库工厂管理器
type FactoryManager struct {
	drivers map[DBType]Driver
}

// NewFactoryManager 创建工厂管理器
func NewFactoryManager() *FactoryManager {
	return &FactoryManager{
		drivers: make(map[DBType]Driver),
	}
}

// Register 注册数据库驱动
func (fm *FactoryManager) Register(driver Driver) {
	fm.drivers[driver.GetDBType()] = driver
}

// GetDriver 获取指定类型的驱动
func (fm *FactoryManager) GetDriver(dbType DBType) (Driver, error) {
	driver, ok := fm.drivers[dbType]
	if !ok {
		return nil, fmt.Errorf("database driver not registered: %w", ErrDBNotSupported)
	}
	return driver, nil
}

// Create 创建数据库连接
func (fm *FactoryManager) Create(dbType DBType, dsn string) (*gorm.DB, error) {
	driver, err := fm.GetDriver(dbType)
	if err != nil {
		return nil, err
	}
	return driver.Open(dsn)
}

// ===== 全局工厂管理器实例 =====

var defaultDBManager *FactoryManager

// InitDB 初始化数据库工厂管理器
func InitDB() *FactoryManager {
	defaultDBManager = NewFactoryManager()

	// 注册默认的驱动
	defaultDBManager.Register(NewMySQLDriver())
	defaultDBManager.Register(NewPostgreSQLDriver())

	return defaultDBManager
}

// GetDBManager 获取默认的数据库工厂管理器
func GetDBManager() *FactoryManager {
	if defaultDBManager == nil {
		return InitDB()
	}
	return defaultDBManager
}

// ===== 便捷构造函数 =====

// NewDB 创建一个数据库实例
// 使用示例:
//
//	db, err := mq.NewDB(context.Background(),
//	    database.WithDriver(database.DBTypeMySQL),
//	    database.WithDSN("user:password@tcp(localhost:3306)/dbname"))
func NewDB(ctx context.Context, opts ...Option) (*DB, error) {
	return New(ctx, opts...)
}

// CreateWithCommonOpts 使用通用选项创建数据库
// 使用示例:
//
//	db, err := database.CreateWithCommonOpts(
//	    database.WithCommonDBType(database.DBTypeMySQL),
//	    database.WithCommonHost("localhost"),
//	    database.WithCommonPort(3306),
//	    database.WithCommonUser("root"),
//	    database.WithCommonPassword("password"),
//	    database.WithCommonDatabase("testdb"))
func CreateWithCommonOpts(ctx context.Context, opts ...CommonOption) (*DB, error) {
	options := &CommonOptions{
		DBType:          DBTypeMySQL,
		Host:            "localhost",
		Port:            3306,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 0,
	}

	for _, opt := range opts {
		opt(options)
	}

	var dsn string
	var driver DBType

	switch options.DBType {
	case DBTypeMySQL:
		dsn = MySQLDSN(options.Host, options.Port, options.User, options.Password, options.Database)
		driver = DBTypeMySQL
	case DBTypePostgres:
		dsn = PostgreSQLDSN(options.Host, options.Port, options.User, options.Password, options.Database)
		driver = DBTypePostgres
	default:
		return nil, ErrDBNotSupported
	}

	return New(ctx,
		WithDriver(driver),
		WithDSN(dsn),
		WithMaxIdleConns(options.MaxIdleConns),
		WithMaxOpenConns(options.MaxOpenConns),
		WithConnMaxLifetime(options.ConnMaxLifetime),
		WithTablePrefix(options.TablePrefix),
	)
}

// ===== 数据库工具函数 =====

// DetectDBType 检测 DSN 对应的数据库类型
func DetectDBType(dsn string) DBType {
	if IsMySQLDSN(dsn) {
		return DBTypeMySQL
	}
	if IsPostgreSQLDSN(dsn) {
		return DBTypePostgres
	}
	return DBTypeMySQL // 默认值
}

// CreateAuto 根据 DSN 自动创建数据库连接
func CreateAuto(ctx context.Context, dsn string, opts ...Option) (*DB, error) {
	dbType := DetectDBType(dsn)

	newOpts := append([]Option{WithDriver(dbType), WithDSN(dsn)}, opts...)
	return New(ctx, newOpts...)
}

// ===== 数据库连接池配置 =====

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

// DefaultPoolConfig 默认连接池配置
var DefaultPoolConfig = PoolConfig{
	MaxIdleConns:    10,
	MaxOpenConns:    100,
	ConnMaxLifetime: 3600, // 秒
	ConnMaxIdleTime: 1800, // 秒
}

// ApplyPoolConfig 应用连接池配置
func ApplyPoolConfig(db *gorm.DB, config *PoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	return nil
}

// ===== 示例代码 =====

// ExampleUsage 示例用法
// package main
//
// import (
//     "context"
//     "github.com/jiujuan/go-ants/pkg/database"
// )
//
// func main() {
//     // 方式1：使用通用选项创建 MySQL
//     db, err := database.CreateWithCommonOpts(context.Background(),
//         database.WithCommonDBType(database.DBTypeMySQL),
//         database.WithCommonHost("localhost"),
//         database.WithCommonPort(3306),
//         database.WithCommonUser("root"),
//         database.WithCommonPassword("password"),
//         database.WithCommonDatabase("testdb"))
//     if err != nil {
//         panic(err)
//     }
//     defer db.Close()
//
//     // 方式2：使用 DSN 创建 PostgreSQL
//     dsn := database.PostgreSQLDSN("localhost", 5432, "postgres", "password", "testdb")
//     db2, err := database.CreateAuto(context.Background(), dsn)
//     if err != nil {
//         panic(err)
//     }
//     defer db2.Close()
//
//     // 方式3：使用选项函数创建
//     db3, err := database.New(context.Background(),
//         database.WithDriver(database.DBTypeMySQL),
//         database.WithDSN("root:password@tcp(localhost:3306)/testdb?charset=utf8mb4"),
//         database.WithMaxOpenConns(50),
//         database.WithMaxIdleConns(10))
//     if err != nil {
//         panic(err)
//     }
//     defer db3.Close()
// }
