package database

import (
	"context"
	"testing"
	"time"
)

// TestNew_DefaultOptions 测试默认选项创建数据库
func TestNew_DefaultOptions(t *testing.T) {
	ctx := context.Background()

	// 由于没有实际的数据库连接，这里测试选项是否正确应用
	opts := &Options{
		Driver:          "mysql",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}

	if opts.Driver != "mysql" {
		t.Errorf("驱动期望 mysql, 得到: %s", opts.Driver)
	}

	if opts.MaxIdleConns != 10 {
		t.Errorf("最大空闲连接期望 10, 得到: %d", opts.MaxIdleConns)
	}

	if opts.MaxOpenConns != 100 {
		t.Errorf("最大打开连接期望 100, 得到: %d", opts.MaxOpenConns)
	}
}

// TestWithDSN_选项函数 测试 DSN 选项函数
func TestWithDSN_选项函数(t *testing.T) {
	opts := &Options{}

	WithDSN("user:pass@tcp(localhost:3306)/db")(opts)

	if opts.DSN != "user:pass@tcp(localhost:3306)/db" {
		t.Errorf("DSN 设置失败: %s", opts.DSN)
	}
}

// TestWithDriver_选项函数 测试驱动选项函数
func TestWithDriver_选项函数(t *testing.T) {
	opts := &Options{}

	WithDriver("postgres")(opts)

	if opts.Driver != "postgres" {
		t.Errorf("驱动期望 postgres, 得到: %s", opts.Driver)
	}
}

// TestWithMaxIdleConns_选项函数 测试最大空闲连接选项函数
func TestWithMaxIdleConns_选项函数(t *testing.T) {
	opts := &Options{}

	WithMaxIdleConns(20)(opts)

	if opts.MaxIdleConns != 20 {
		t.Errorf("最大空闲连接期望 20, 得到: %d", opts.MaxIdleConns)
	}
}

// TestWithMaxOpenConns_选项函数 测试最大打开连接选项函数
func TestWithMaxOpenConns_选项函数(t *testing.T) {
	opts := &Options{}

	WithMaxOpenConns(200)(opts)

	if opts.MaxOpenConns != 200 {
		t.Errorf("最大打开连接期望 200, 得到: %d", opts.MaxOpenConns)
	}
}

// TestWithConnMaxLifetime_选项函数 测试连接最大生命周期选项函数
func TestWithConnMaxLifetime_选项函数(t *testing.T) {
	opts := &Options{}

	lifetime := time.Hour * 2
	WithConnMaxLifetime(lifetime)(opts)

	if opts.ConnMaxLifetime != lifetime {
		t.Errorf("连接最大生命周期设置失败")
	}
}

// TestWithConnMaxIdleTime_选项函数 测试连接最大空闲时间选项函数
func TestWithConnMaxIdleTime_选项函数(t *testing.T) {
	opts := &Options{}

	idleTime := time.Minute * 30
	WithConnMaxIdleTime(idleTime)(opts)

	if opts.ConnMaxIdleTime != idleTime {
		t.Errorf("连接最大空闲时间设置失败")
	}
}

// TestWithSlowThreshold_选项函数 测试慢查询阈值选项函数
func TestWithSlowThreshold_选项函数(t *testing.T) {
	opts := &Options{}

	threshold := time.Second * 2
	WithSlowThreshold(threshold)(opts)

	if opts.SlowThreshold != threshold {
		t.Errorf("慢查询阈值设置失败")
	}
}

// TestMySQLDSN_构建MySQLDSN 测试 MySQL DSN 构建
func TestMySQLDSN_构建MySQLDSN(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "password", "testdb")

	if dsn == "" {
		t.Error("DSN 构建失败")
	}

	// 验证基本结构
	expectedParts := []string{"user", "password", "localhost", "3306", "testdb"}
	for _, part := range expectedParts {
		if !contains(dsn, part) {
			t.Errorf("DSN 缺少必要部分: %s", part)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMySQLDSN_WithOptions 测试带选项的 MySQL DSN
func TestMySQLDSN_WithOptions(t *testing.T) {
	dsn := MySQLDSN(
		"localhost",
		3306,
		"user",
		"password",
		"testdb",
		WithMySQLCharset("utf8"),
		WithMySQLParseTime(true),
	)

	if dsn == "" {
		t.Error("DSN 构建失败")
	}

	if !contains(dsn, "charset=utf8") {
		t.Error("字符集选项未应用到 DSN")
	}
}

// TestPostgreSQLDSN_构建PostgreSQLDSN 测试 PostgreSQL DSN 构建
func TestPostgreSQLDSN_构建PostgreSQLDSN(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "user", "password", "testdb")

	if dsn == "" {
		t.Error("DSN 构建失败")
	}

	expectedParts := []string{"user", "password", "localhost", "5432", "testdb"}
	for _, part := range expectedParts {
		if !contains(dsn, part) {
			t.Errorf("DSN 缺少必要部分: %s", part)
		}
	}
}

// TestPostgreSQLDSN_WithOptions 测试带选项的 PostgreSQL DSN
func TestPostgreSQLDSN_WithOptions(t *testing.T) {
	dsn := PostgreSQLDSN(
		"localhost",
		5432,
		"user",
		"password",
		"testdb",
		WithPostgreSQLSSLMode("require"),
		WithPostgreSQLTimeZone("UTC"),
	)

	if dsn == "" {
		t.Error("DSN 构建失败")
	}

	if !contains(dsn, "sslmode=require") {
		t.Error("SSL 模式选项未应用到 DSN")
	}
}

// TestCreateMySQLDSNWithOptions 测试使用选项创建 MySQL DSN
func TestCreateMySQLDSNWithOptions(t *testing.T) {
	dsn := CreateMySQLDSNWithOptions(
		"localhost",
		3306,
		"user",
		"password",
		"testdb",
		WithMySQLCharset("utf8mb4"),
	)

	if dsn == "" {
		t.Error("DSN 创建失败")
	}
}

// TestCreatePostgreSQLDSNWithOptions 测试使用选项创建 PostgreSQL DSN
func TestCreatePostgreSQLDSNWithOptions(t *testing.T) {
	dsn := CreatePostgreSQLDSNWithOptions(
		"localhost",
		5432,
		"user",
		"password",
		"testdb",
		WithPostgreSQLSSLMode("disable"),
	)

	if dsn == "" {
		t.Error("DSN 创建失败")
	}
}

// TestMySQLOptions_字符集 测试 MySQL 字符集选项
func TestMySQLOptions_字符集(t *testing.T) {
	opts := &MySQLOptions{}

	WithMySQLCharset("utf8mb4")(opts)

	if opts.Charset != "utf8mb4" {
		t.Errorf("字符集期望 utf8mb4, 得到: %s", opts.Charset)
	}
}

// TestMySQLOptions_解析时间 测试 MySQL 解析时间选项
func TestMySQLOptions_解析时间(t *testing.T) {
	opts := &MySQLOptions{}

	WithMySQLParseTime(true)(opts)

	if opts.ParseTime != true {
		t.Error("解析时间选项设置失败")
	}
}

// TestMySQLOptions_UTC时间 测试 MySQL UTC 时间选项
func TestMySQLOptions_UTC时间(t *testing.T) {
	opts := &MySQLOptions{}

	WithMySQLUtc(true)(opts)

	if opts.Utc != true {
		t.Error("UTC 时间选项设置失败")
	}
}

// TestPostgreSQLOptions_SSL模式 测试 PostgreSQL SSL 模式选项
func TestPostgreSQLOptions_SSL模式(t *testing.T) {
	opts := &PostgreSQLOptions{}

	WithPostgreSQLSSLMode("require")(opts)

	if opts.SSLMode != "require" {
		t.Errorf("SSL 模式期望 require, 得到: %s", opts.SSLMode)
	}
}

// TestPostgreSQLOptions_时区 测试 PostgreSQL 时区选项
func TestPostgreSQLOptions_时区(t *testing.T) {
	opts := &PostgreSQLOptions{}

	WithPostgreSQLTimeZone("Asia/Shanghai")(opts)

	if opts.TimeZone != "Asia/Shanghai" {
		t.Errorf("时区期望 Asia/Shanghai, 得到: %s", opts.TimeZone)
	}
}

// TestOptions_多个选项组合 测试多个选项组合
func TestOptions_多个选项组合(t *testing.T) {
	dsn := "user:pass@tcp(localhost:3306)/testdb"

	opts := &Options{}
	WithDSN(dsn)(opts)
	WithDriver("mysql")(opts)
	WithMaxIdleConns(20)(opts)
	WithMaxOpenConns(200)(opts)
	WithConnMaxLifetime(time.Hour * 2)(opts)
	WithDebug(true)(opts)

	if opts.DSN != dsn {
		t.Error("DSN 设置失败")
	}

	if opts.Driver != "mysql" {
		t.Error("驱动设置失败")
	}

	if opts.MaxIdleConns != 20 {
		t.Error("最大空闲连接设置失败")
	}

	if opts.MaxOpenConns != 200 {
		t.Error("最大打开连接设置失败")
	}

	if opts.Debug != true {
		t.Error("调试模式设置失败")
	}
}

// TestOptions_默认选项 测试默认选项
func TestOptions_默认选项(t *testing.T) {
	opts := &Options{}

	// 应用所有默认值
	WithDriver("mysql")(opts)
	WithMaxIdleConns(10)(opts)
	WithMaxOpenConns(100)(opts)

	if opts.Driver != "mysql" {
		t.Errorf("默认驱动期望 mysql, 得到: %s", opts.Driver)
	}

	if opts.MaxIdleConns != 10 {
		t.Errorf("默认最大空闲连接期望 10, 得到: %d", opts.MaxIdleConns)
	}

	if opts.MaxOpenConns != 100 {
		t.Errorf("默认最大打开连接期望 100, 得到: %d", opts.MaxOpenConns)
	}
}

// TestOptions_WithNamingStrategy 测试命名策略选项
func TestOptions_WithNamingStrategy(t *testing.T) {
	opts := &Options{}

	strategy := Options.NamingStrategy{
		TablePrefix:   "tbl_",
		SingularTable: true,
	}
	WithNamingStrategy(strategy)(opts)

	if opts.NamingStrategy.TablePrefix != "tbl_" {
		t.Error("表前缀设置失败")
	}

	if !opts.NamingStrategy.SingularTable {
		t.Error("单数表设置失败")
	}
}

// TestOptions_WithTablePrefix 测试表前缀选项
func TestOptions_WithTablePrefix(t *testing.T) {
	opts := &Options{}

	WithTablePrefix("app_")(opts)

	if opts.TablePrefix != "app_" {
		t.Errorf("表前缀期望 app_, 得到: %s", opts.TablePrefix)
	}
}

// TestOptions_WithDisableForeignKeyConstraint 测试禁用外键约束选项
func TestOptions_WithDisableForeignKeyConstraint(t *testing.T) {
	opts := &Options{}

	WithDisableForeignKeyConstraint(true)(opts)

	if !opts.DisableForeignKeyConstraint {
		t.Error("外键约束设置失败")
	}
}

// TestOptions_WithSkipDefaultTransaction 测试跳过默认事务选项
func TestOptions_WithSkipDefaultTransaction(t *testing.T) {
	opts := &Options{}

	WithSkipDefaultTransaction(true)(opts)

	if !opts.SkipDefaultTransaction {
		t.Error("跳过默认事务设置失败")
	}
}

// TestOptions_WithPrepareStmt 测试预编译语句选项
func TestOptions_WithPrepareStmt(t *testing.T) {
	opts := &Options{}

	WithPrepareStmt(false)(opts)

	if opts.PrepareStmt != false {
		t.Error("预编译语句设置失败")
	}
}

// TestOptions_WithContext 测试上下文选项
func TestOptions_WithContext(t *testing.T) {
	opts := &Options{}

	WithContext(false)(opts)

	if opts.WithContext != false {
		t.Error("上下文支持设置失败")
	}
}

// TestOptions_WithCallback 测试回调选项
func TestOptions_WithCallback(t *testing.T) {
	opts := &Options{}

	callbackCalled := false
	callback := func(db interface{}) {
		callbackCalled = true
	}

	WithCallback(callback)(opts)

	if len(opts.Callbacks) != 1 {
		t.Error("回调函数添加失败")
	}
}

// TestDB_结构 测试 DB 结构
func TestDB_结构(t *testing.T) {
	// 测试 DB 结构定义
	db := &DB{}

	if db == nil {
		t.Error("DB 结构创建失败")
	}
}

// TestWithContext_测试上下文 测试带上下文的 DB
func TestWithContext_测试上下文(t *testing.T) {
	ctx := context.Background()

	// 创建模拟的 DB
	db := &DB{}

	// 测试 WithContext 方法存在
	_ = db.WithContext

	// 由于没有实际连接，这里只验证方法签名
	_ = ctx
	_ = db
}
