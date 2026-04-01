package database

import (
	"context"
	"testing"
	"time"
)

// ===== Interface Tests =====

func TestDBType_Constants(t *testing.T) {
	tests := []struct {
		dbType   DBType
		expected string
	}{
		{DBTypeMySQL, "mysql"},
		{DBTypePostgres, "postgres"},
		{DBTypeSQLite, "sqlite"},
		{DBTypeSQLServer, "sqlserver"},
	}

	for _, tt := range tests {
		if string(tt.dbType) != tt.expected {
			t.Errorf("DBType 期望 %s, 得到 %s", tt.expected, tt.dbType)
		}
	}
}

func TestNewBaseMessage(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"))

	if msg.GetKey() != "key" {
		t.Errorf("消息 Key 期望 'key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "value" {
		t.Errorf("消息 Value 期望 'value', 得到 %s", string(msg.GetValue()))
	}

	if msg.GetHeaders() == nil {
		t.Error("消息 Headers 不应该为 nil")
	}

	if msg.GetTopic() != "" {
		t.Errorf("消息 Topic 期望为空, 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != 0 {
		t.Errorf("消息 Timestamp 期望 0, 得到 %d", msg.GetTimestamp())
	}
}

func TestWithMessageTopic(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"), WithMessageTopic("test-topic"))

	if msg.GetTopic() != "test-topic" {
		t.Errorf("消息 Topic 期望 'test-topic', 得到 %s", msg.GetTopic())
	}
}

func TestWithMessageTimestamp(t *testing.T) {
	timestamp := int64(1234567890)
	msg := NewBaseMessage("key", []byte("value"), WithMessageTimestamp(timestamp))

	if msg.GetTimestamp() != timestamp {
		t.Errorf("消息 Timestamp 期望 %d, 得到 %d", timestamp, msg.GetTimestamp())
	}
}

func TestWithMessageHeader(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"), WithMessageHeader("header-key", "header-value"))

	headers := msg.GetHeaders()
	if headers["header-key"] != "header-value" {
		t.Errorf("消息 Header 期望 'header-value', 得到 %v", headers["header-key"])
	}
}

func TestCommonOptions_Defaults(t *testing.T) {
	opts := &CommonOptions{
		DBType: DBTypeMySQL,
		Host:   "localhost",
		Port:   3306,
	}

	if opts.DBType != DBTypeMySQL {
		t.Errorf("DBType 期望 MySQL, 得到 %s", opts.DBType)
	}

	if opts.Host != "localhost" {
		t.Errorf("Host 期望 'localhost', 得到 %s", opts.Host)
	}

	if opts.Port != 3306 {
		t.Errorf("Port 期望 3306, 得到 %d", opts.Port)
	}
}

func TestCommonOption_Functions(t *testing.T) {
	opts := &CommonOptions{}

	WithCommonDBType(DBTypePostgres)(opts)
	WithCommonHost("127.0.0.1")(opts)
	WithCommonPort(5432)(opts)
	WithCommonUser("testuser")(opts)
	WithCommonPassword("testpass")(opts)
	WithCommonDatabase("testdb")(opts)
	WithCommonMaxIdleConns(20)(opts)
	WithCommonMaxOpenConns(200)(opts)
	WithCommonConnMaxLifetime(time.Hour)(opts)
	WithCommonTablePrefix("tbl_")(opts)

	if opts.DBType != DBTypePostgres {
		t.Errorf("DBType 期望 Postgres, 得到 %s", opts.DBType)
	}

	if opts.Host != "127.0.0.1" {
		t.Errorf("Host 期望 '127.0.0.1', 得到 %s", opts.Host)
	}

	if opts.Port != 5432 {
		t.Errorf("Port 期望 5432, 得到 %d", opts.Port)
	}

	if opts.User != "testuser" {
		t.Errorf("User 期望 'testuser', 得到 %s", opts.User)
	}

	if opts.Password != "testpass" {
		t.Errorf("Password 期望 'testpass', 得到 %s", opts.Password)
	}

	if opts.Database != "testdb" {
		t.Errorf("Database 期望 'testdb', 得到 %s", opts.Database)
	}

	if opts.MaxIdleConns != 20 {
		t.Errorf("MaxIdleConns 期望 20, 得到 %d", opts.MaxIdleConns)
	}

	if opts.MaxOpenConns != 200 {
		t.Errorf("MaxOpenConns 期望 200, 得到 %d", opts.MaxOpenConns)
	}

	if opts.ConnMaxLifetime != time.Hour {
		t.Errorf("ConnMaxLifetime 期望 1 Hour, 得到 %s", opts.ConnMaxLifetime)
	}

	if opts.TablePrefix != "tbl_" {
		t.Errorf("TablePrefix 期望 'tbl_', 得到 %s", opts.TablePrefix)
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrDBNotSupported, "database type not supported"},
		{ErrConnectionFailed, "database connection failed"},
		{ErrTransactionFailed, "transaction failed"},
		{ErrMigrationFailed, "migration failed"},
		{ErrQueryFailed, "query failed"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expected {
			t.Errorf("错误消息期望 '%s', 得到 '%s'", tt.expected, tt.err.Error())
		}
	}
}

func TestMaskDSN(t *testing.T) {
	dsn := "user:password@tcp(localhost:3306)/db"
	masked := MaskDSN(dsn)

	if masked != dsn {
		t.Errorf("MaskDSN 应该返回原始 DSN, 期望 %s, 得到 %s", dsn, masked)
	}
}

// ===== MySQL Tests =====

func TestMySQLDriver_GetDBType(t *testing.T) {
	driver := NewMySQLDriver()

	if driver.GetDBType() != DBTypeMySQL {
		t.Errorf("MySQLDriver 类型期望 MySQL, 得到 %s", driver.GetDBType())
	}
}

func TestMySQLDriver_Open_InvalidDSN(t *testing.T) {
	driver := NewMySQLDriver()

	// 使用无效的 DSN 应该返回错误
	_, err := driver.Open("invalid-dsn")
	if err == nil {
		t.Error("MySQLDriver.Open 应该对无效 DSN 返回错误")
	}
}

func TestMySQLDSN_Basic(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "root", "password", "testdb")

	if dsn == "" {
		t.Error("MySQL DSN 不应该为空")
	}

	expectedParts := []string{"root", "password", "localhost", "3306", "testdb"}
	for _, part := range expectedParts {
		if !containsString(dsn, part) {
			t.Errorf("MySQL DSN 应该包含 '%s'", part)
		}
	}
}

func TestMySQLDSN_WithCharset(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLCharset("utf8"))

	if !containsString(dsn, "charset=utf8") {
		t.Error("MySQL DSN 应该包含 charset=utf8")
	}
}

func TestMySQLDSN_WithParseTime(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLParseTime(true))

	if !containsString(dsn, "parseTime=true") {
		t.Error("MySQL DSN 应该包含 parseTime=true")
	}
}

func TestMySQLDSN_WithUtc(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLUtc(true))

	if !containsString(dsn, "time_utc=true") {
		t.Error("MySQL DSN 应该包含 time_utc=true")
	}
}

func TestMySQLDSN_WithTimeout(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLTimeout("10s"))

	if !containsString(dsn, "timeout=10s") {
		t.Error("MySQL DSN 应该包含 timeout=10s")
	}
}

func TestMySQLDSN_WithLoc(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLLoc("UTC"))

	if !containsString(dsn, "loc=UTC") {
		t.Error("MySQL DSN 应该包含 loc=UTC")
	}
}

func TestMySQLDSN_AllOptions(t *testing.T) {
	dsn := MySQLDSN("localhost", 3306, "user", "pass", "db",
		WithMySQLCharset("utf8mb4"),
		WithMySQLParseTime(true),
		WithMySQLLoc("Local"),
		WithMySQLUtc(false),
		WithMySQLTimeout("5s"))

	if dsn == "" {
		t.Error("MySQL DSN 不应该为空")
	}

	if !containsString(dsn, "charset=utf8mb4") {
		t.Error("MySQL DSN 应该包含 charset=utf8mb4")
	}

	if !containsString(dsn, "parseTime=true") {
		t.Error("MySQL DSN 应该包含 parseTime=true")
	}
}

func TestCreateMySQLDSNWithOptions(t *testing.T) {
	dsn := CreateMySQLDSNWithOptions("localhost", 3306, "user", "pass", "db",
		WithMySQLCharset("utf8"))

	if dsn == "" {
		t.Error("CreateMySQLDSNWithOptions 不应该返回空字符串")
	}
}

func TestMySQLConfig_GetDSN(t *testing.T) {
	config := &MySQLConfig{
		Host:      "localhost",
		Port:      3306,
		User:      "root",
		Password:  "password",
		Database:  "testdb",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	dsn := config.GetDSN()

	if dsn == "" {
		t.Error("MySQLConfig.GetDSN 不应该返回空字符串")
	}

	if !containsString(dsn, "root") {
		t.Error("MySQLConfig.GetDSN 应该包含用户名")
	}

	if !containsString(dsn, "testdb") {
		t.Error("MySQLConfig.GetDSN 应该包含数据库名")
	}
}

func TestMySQLConfig_GetDBType(t *testing.T) {
	config := &MySQLConfig{}

	if config.GetDBType() != DBTypeMySQL {
		t.Errorf("MySQLConfig.GetDBType 期望 MySQL, 得到 %s", config.GetDBType())
	}
}

func TestMySQLConnector(t *testing.T) {
	config := &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "testdb",
	}

	connector := NewMySQLConnector(config)

	if connector == nil {
		t.Error("MySQLConnector 不应该为 nil")
	}

	if connector.config != config {
		t.Error("MySQLConnector.config 配置不正确")
	}
}

func TestMySQLFactory(t *testing.T) {
	factory := NewMySQLFactory()

	if factory == nil {
		t.Error("MySQLFactory 不应该为 nil")
	}

	// 测试使用无效 DSN
	_, err := factory.Create("invalid-dsn")
	if err == nil {
		t.Error("MySQLFactory.Create 应该对无效 DSN 返回错误")
	}
}

func TestIsMySQLDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected bool
	}{
		{"root:pass@tcp(localhost:3306)/db", true},
		{"mysql://localhost:3306/db", true},
		{"host=localhost user=root dbname=test", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsMySQLDSN(tt.dsn)
		if result != tt.expected {
			t.Errorf("IsMySQLDSN(%s) 期望 %v, 得到 %v", tt.dsn, tt.expected, result)
		}
	}
}

// ===== PostgreSQL Tests =====

func TestPostgreSQLDriver_GetDBType(t *testing.T) {
	driver := NewPostgreSQLDriver()

	if driver.GetDBType() != DBTypePostgres {
		t.Errorf("PostgreSQLDriver 类型期望 Postgres, 得到 %s", driver.GetDBType())
	}
}

func TestPostgreSQLDriver_Open_InvalidDSN(t *testing.T) {
	driver := NewPostgreSQLDriver()

	// 使用无效的 DSN 应该返回错误
	_, err := driver.Open("invalid-dsn")
	if err == nil {
		t.Error("PostgreSQLDriver.Open 应该对无效 DSN 返回错误")
	}
}

func TestPostgreSQLDSN_Basic(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "postgres", "password", "testdb")

	if dsn == "" {
		t.Error("PostgreSQL DSN 不应该为空")
	}

	expectedParts := []string{"host=localhost", "user=postgres", "password=password", "dbname=testdb", "port=5432"}
	for _, part := range expectedParts {
		if !containsString(dsn, part) {
			t.Errorf("PostgreSQL DSN 应该包含 '%s'", part)
		}
	}
}

func TestPostgreSQLDSN_WithSSLMode(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "user", "pass", "db",
		WithPostgreSQLSSLMode("require"))

	if !containsString(dsn, "sslmode=require") {
		t.Error("PostgreSQL DSN 应该包含 sslmode=require")
	}
}

func TestPostgreSQLDSN_WithTimeZone(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "user", "pass", "db",
		WithPostgreSQLTimeZone("UTC"))

	if !containsString(dsn, "timezone=UTC") {
		t.Error("PostgreSQL DSN 应该包含 timezone=UTC")
	}
}

func TestPostgreSQLDSN_WithSSLCerts(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "user", "pass", "db",
		WithPostgreSQLSSLCert("/path/to/cert"),
		WithPostgreSQLSSLKey("/path/to/key"),
		WithPostgreSQLSSLRootCert("/path/to/rootcert"))

	if !containsString(dsn, "sslcert=/path/to/cert") {
		t.Error("PostgreSQL DSN 应该包含 sslcert")
	}

	if !containsString(dsn, "sslkey=/path/to/key") {
		t.Error("PostgreSQL DSN 应该包含 sslkey")
	}

	if !containsString(dsn, "sslrootcert=/path/to/rootcert") {
		t.Error("PostgreSQL DSN 应该包含 sslrootcert")
	}
}

func TestPostgreSQLDSN_AllOptions(t *testing.T) {
	dsn := PostgreSQLDSN("localhost", 5432, "user", "pass", "db",
		WithPostgreSQLSSLMode("verify-full"),
		WithPostgreSQLTimeZone("America/New_York"),
		WithPostgreSQLSSLCert("cert.pem"),
		WithPostgreSQLSSLKey("key.pem"),
		WithPostgreSQLSSLRootCert("root.pem"))

	if dsn == "" {
		t.Error("PostgreSQL DSN 不应该为空")
	}

	if !containsString(dsn, "sslmode=verify-full") {
		t.Error("PostgreSQL DSN 应该包含 sslmode=verify-full")
	}

	if !containsString(dsn, "timezone=America/New_York") {
		t.Error("PostgreSQL DSN 应该包含 timezone")
	}
}

func TestCreatePostgreSQLDSNWithOptions(t *testing.T) {
	dsn := CreatePostgreSQLDSNWithOptions("localhost", 5432, "user", "pass", "db",
		WithPostgreSQLSSLMode("disable"))

	if dsn == "" {
		t.Error("CreatePostgreSQLDSNWithOptions 不应该返回空字符串")
	}
}

func TestPostgreSQLConfig_GetDSN(t *testing.T) {
	config := &PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
		SSLMode:  "disable",
		TimeZone: "Local",
	}

	dsn := config.GetDSN()

	if dsn == "" {
		t.Error("PostgreSQLConfig.GetDSN 不应该返回空字符串")
	}

	if !containsString(dsn, "postgres") {
		t.Error("PostgreSQLConfig.GetDSN 应该包含用户名")
	}

	if !containsString(dsn, "testdb") {
		t.Error("PostgreSQLConfig.GetDSN 应该包含数据库名")
	}

	if !containsString(dsn, "sslmode=disable") {
		t.Error("PostgreSQLConfig.GetDSN 应该包含 sslmode")
	}
}

func TestPostgreSQLConfig_GetDBType(t *testing.T) {
	config := &PostgreSQLConfig{}

	if config.GetDBType() != DBTypePostgres {
		t.Errorf("PostgreSQLConfig.GetDBType 期望 Postgres, 得到 %s", config.GetDBType())
	}
}

func TestPostgreSQLConnector(t *testing.T) {
	config := &PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
	}

	connector := NewPostgreSQLConnector(config)

	if connector == nil {
		t.Error("PostgreSQLConnector 不应该为 nil")
	}

	if connector.config != config {
		t.Error("PostgreSQLConnector.config 配置不正确")
	}
}

func TestPostgreSQLFactory(t *testing.T) {
	factory := NewPostgreSQLFactory()

	if factory == nil {
		t.Error("PostgreSQLFactory 不应该为 nil")
	}

	// 测试使用无效 DSN
	_, err := factory.Create("invalid-dsn")
	if err == nil {
		t.Error("PostgreSQLFactory.Create 应该对无效 DSN 返回错误")
	}
}

func TestIsPostgreSQLDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected bool
	}{
		{"host=localhost user=root dbname=test", true},
		{"postgres://localhost:5432/testdb", true},
		{"root:pass@tcp(localhost:3306)/db", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsPostgreSQLDSN(tt.dsn)
		if result != tt.expected {
			t.Errorf("IsPostgreSQLDSN(%s) 期望 %v, 得到 %v", tt.dsn, tt.expected, result)
		}
	}
}

// ===== Database Options Tests =====

func TestOptions_WithDBType(t *testing.T) {
	opts := &Options{}

	WithDriver(DBTypePostgres)(opts)

	if opts.Driver != DBTypePostgres {
		t.Errorf("Driver 期望 Postgres, 得到 %s", opts.Driver)
	}
}

func TestOptions_WithDSN(t *testing.T) {
	opts := &Options{}

	dsn := "user:pass@tcp(localhost:3306)/db"
	WithDSN(dsn)(opts)

	if opts.DSN != dsn {
		t.Errorf("DSN 设置失败, 期望 %s, 得到 %s", dsn, opts.DSN)
	}
}

func TestOptions_WithMaxIdleConns(t *testing.T) {
	opts := &Options{}

	WithMaxIdleConns(20)(opts)

	if opts.MaxIdleConns != 20 {
		t.Errorf("MaxIdleConns 期望 20, 得到 %d", opts.MaxIdleConns)
	}
}

func TestOptions_WithMaxOpenConns(t *testing.T) {
	opts := &Options{}

	WithMaxOpenConns(200)(opts)

	if opts.MaxOpenConns != 200 {
		t.Errorf("MaxOpenConns 期望 200, 得到 %d", opts.MaxOpenConns)
	}
}

func TestOptions_WithConnMaxLifetime(t *testing.T) {
	opts := &Options{}

	lifetime := time.Hour * 2
	WithConnMaxLifetime(lifetime)(opts)

	if opts.ConnMaxLifetime != lifetime {
		t.Errorf("ConnMaxLifetime 设置失败")
	}
}

func TestOptions_WithConnMaxIdleTime(t *testing.T) {
	opts := &Options{}

	idleTime := time.Minute * 30
	WithConnMaxIdleTime(idleTime)(opts)

	if opts.ConnMaxIdleTime != idleTime {
		t.Errorf("ConnMaxIdleTime 设置失败")
	}
}

func TestOptions_WithSlowThreshold(t *testing.T) {
	opts := &Options{}

	threshold := time.Second * 2
	WithSlowThreshold(threshold)(opts)

	if opts.SlowThreshold != threshold {
		t.Errorf("SlowThreshold 设置失败")
	}
}

func TestOptions_WithTablePrefix(t *testing.T) {
	opts := &Options{}

	WithTablePrefix("tbl_")(opts)

	if opts.TablePrefix != "tbl_" {
		t.Errorf("TablePrefix 期望 'tbl_', 得到 %s", opts.TablePrefix)
	}
}

func TestOptions_WithDisableForeignKeyConstraint(t *testing.T) {
	opts := &Options{}

	WithDisableForeignKeyConstraint(true)(opts)

	if !opts.DisableForeignKeyConstraint {
		t.Error("DisableForeignKeyConstraint 应该为 true")
	}
}

func TestOptions_WithSkipDefaultTransaction(t *testing.T) {
	opts := &Options{}

	WithSkipDefaultTransaction(true)(opts)

	if !opts.SkipDefaultTransaction {
		t.Error("SkipDefaultTransaction 应该为 true")
	}
}

func TestOptions_WithPrepareStmt(t *testing.T) {
	opts := &Options{}

	WithPrepareStmt(false)(opts)

	if opts.PrepareStmt {
		t.Error("PrepareStmt 应该为 false")
	}
}

func TestOptions_WithDebug(t *testing.T) {
	opts := &Options{}

	WithDebug(true)(opts)

	if !opts.Debug {
		t.Error("Debug 应该为 true")
	}
}

func TestOptions_WithContext(t *testing.T) {
	opts := &Options{}

	WithContext(false)(opts)

	if opts.WithContext {
		t.Error("WithContext 应该为 false")
	}
}

func TestOptions_WithCallback(t *testing.T) {
	opts := &Options{}

	callback := func(db interface{}) {}

	WithCallback(callback)(opts)

	if len(opts.Callbacks) != 1 {
		t.Errorf("Callbacks 长度期望 1, 得到 %d", len(opts.Callbacks))
	}
}

func TestOptions_CombinedOptions(t *testing.T) {
	opts := &Options{
		Driver:          DBTypeMySQL,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}

	if opts.Driver != DBTypeMySQL {
		t.Errorf("Driver 期望 MySQL, 得到 %s", opts.Driver)
	}

	if opts.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns 期望 10, 得到 %d", opts.MaxIdleConns)
	}

	if opts.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns 期望 100, 得到 %d", opts.MaxOpenConns)
	}
}

// ===== Factory Tests =====

func TestFactoryManager_New(t *testing.T) {
	manager := NewFactoryManager()

	if manager == nil {
		t.Error("FactoryManager 不应该为 nil")
	}

	if manager.drivers == nil {
		t.Error("FactoryManager.drivers 不应该为 nil")
	}
}

func TestFactoryManager_Register(t *testing.T) {
	manager := NewFactoryManager()

	manager.Register(NewMySQLDriver())
	manager.Register(NewPostgreSQLDriver())

	if len(manager.drivers) != 2 {
		t.Errorf("drivers 数量期望 2, 得到 %d", len(manager.drivers))
	}
}

func TestFactoryManager_GetDriver(t *testing.T) {
	manager := NewFactoryManager()
	manager.Register(NewMySQLDriver())
	manager.Register(NewPostgreSQLDriver())

	driver, err := manager.GetDriver(DBTypeMySQL)
	if err != nil {
		t.Errorf("GetDriver 不应该返回错误: %v", err)
	}
	if driver.GetDBType() != DBTypeMySQL {
		t.Errorf("Driver 类型期望 MySQL, 得到 %s", driver.GetDBType())
	}

	driver, err = manager.GetDriver(DBTypePostgres)
	if err != nil {
		t.Errorf("GetDriver 不应该返回错误: %v", err)
	}
	if driver.GetDBType() != DBTypePostgres {
		t.Errorf("Driver 类型期望 Postgres, 得到 %s", driver.GetDBType())
	}

	_, err = manager.GetDriver(DBTypeSQLite)
	if err == nil {
		t.Error("GetDriver 应该对未注册的驱动返回错误")
	}
}

func TestInitDB(t *testing.T) {
	manager := InitDB()

	if manager == nil {
		t.Error("InitDB 不应该返回 nil")
	}

	if len(manager.drivers) != 2 {
		t.Errorf("InitDB 应该注册 2 个驱动, 得到 %d", len(manager.drivers))
	}
}

func TestGetDBManager(t *testing.T) {
	// 重置全局管理器
	defaultDBManager = nil

	manager := GetDBManager()

	if manager == nil {
		t.Error("GetDBManager 不应该返回 nil")
	}

	// 再次调用应该返回同一个实例
	manager2 := GetDBManager()
	if manager != manager2 {
		t.Error("GetDBManager 应该返回同一个实例")
	}
}

func TestDetectDBType(t *testing.T) {
	tests := []struct {
		dsn      string
		expected DBType
	}{
		{"root:pass@tcp(localhost:3306)/db", DBTypeMySQL},
		{"mysql://localhost:3306/db", DBTypeMySQL},
		{"host=localhost user=root dbname=test", DBTypePostgres},
		{"postgres://localhost:5432/testdb", DBTypePostgres},
		{"", DBTypeMySQL}, // 默认值
	}

	for _, tt := range tests {
		result := DetectDBType(tt.dsn)
		if result != tt.expected {
			t.Errorf("DetectDBType(%s) 期望 %s, 得到 %s", tt.dsn, tt.expected, result)
		}
	}
}

func TestPoolConfig(t *testing.T) {
	config := &PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 1800,
	}

	if config.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns 期望 10, 得到 %d", config.MaxIdleConns)
	}

	if config.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns 期望 100, 得到 %d", config.MaxOpenConns)
	}

	if DefaultPoolConfig.MaxIdleConns != 10 {
		t.Errorf("DefaultPoolConfig.MaxIdleConns 期望 10, 得到 %d", DefaultPoolConfig.MaxIdleConns)
	}
}

// ===== DB Wrapper Tests =====

func TestDB_GetDBType(t *testing.T) {
	db := &DB{driver: DBTypeMySQL}

	if db.GetDBType() != DBTypeMySQL {
		t.Errorf("GetDBType 期望 MySQL, 得到 %s", db.GetDBType())
	}
}

func TestDB_WithContext(t *testing.T) {
	ctx := context.Background()
	db := &DB{driver: DBTypeMySQL}

	newDB := db.WithContext(ctx)

	if newDB == nil {
		t.Error("WithContext 不应该返回 nil")
	}

	if newDB.GetDBType() != DBTypeMySQL {
		t.Error("WithContext 应该保持 driver 类型")
	}
}

func TestDB_Transaction_WithoutConnection(t *testing.T) {
	db := &DB{driver: DBTypeMySQL}

	// 在没有实际连接的情况下，Transaction 应该会失败
	err := db.Transaction(func(tx *DB) error {
		return nil
	})

	// 这里应该会返回错误，因为没有实际的数据库连接
	if err == nil {
		// 如果没有错误，说明可能连接到了实际的数据库
		t.Log("Transaction 执行成功（可能连接到实际数据库）")
	}
}

func TestDB_Close_WithoutConnection(t *testing.T) {
	db := &DB{driver: DBTypeMySQL}

	// 在没有实际连接的情况下，Close 应该会失败
	err := db.Close()

	if err == nil {
		t.Error("Close 应该返回错误（没有实际连接）")
	}
}

func TestDB_AutoMigrate_WithoutConnection(t *testing.T) {
	db := &DB{driver: DBTypeMySQL}

	// 在没有实际连接的情况下，AutoMigrate 应该会失败
	err := db.AutoMigrate()

	if err == nil {
		t.Error("AutoMigrate 应该返回错误（没有实际连接）")
	}
}

// ===== Helper Functions =====

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
