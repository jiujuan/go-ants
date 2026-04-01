package conf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNew_创建配置实例 测试创建配置实例
func TestNew_创建配置实例(t *testing.T) {
	c := New()
	if c == nil {
		t.Error("配置实例创建失败")
	}
}

// TestNewWithConfig_使用现有Viper 测试使用现有 Viper 实例
func TestNewWithConfig_使用现有Viper(t *testing.T) {
	v := New().GetViper()
	c := NewWithConfig(v)
	if c == nil {
		t.Error("配置实例创建失败")
	}
}

// TestLoad_文件不存在 测试文件不存在的情况
func TestLoad_文件不存在(t *testing.T) {
	c := New()
	err := c.Load("nonexistent.yaml")
	// 应该返回 nil，因为文件不存在且没有设置配置文件
	if err != nil {
		t.Logf("预期 nil, 得到: %v", err)
	}
}

// TestLoad_无效配置 测试无效配置文件
func TestLoad_无效配置(t *testing.T) {
	c := New()
	// 创建一个临时无效配置文件
	tmpFile := filepath.Join(os.TempDir(), "test_invalid.yaml")
	os.WriteFile(tmpFile, []byte("invalid: yaml: content: ["), 0644)
	defer os.Remove(tmpFile)

	err := c.Load(tmpFile)
	if err == nil {
		t.Error("应该返回错误")
	}
}

// TestGet_获取配置值 测试获取配置值
func TestGet_获取配置值(t *testing.T) {
	c := New()
	c.Set("test.key", "test-value")

	value := c.Get("test.key")
	if value != "test-value" {
		t.Errorf("期望 test-value, 得到: %v", value)
	}
}

// TestGetString_获取字符串 测试获取字符串配置
func TestGetString_获取字符串(t *testing.T) {
	c := New()
	c.Set("string.key", "hello")

	value := c.GetString("string.key")
	if value != "hello" {
		t.Errorf("期望 hello, 得到: %s", value)
	}
}

// TestGetInt_获取整数 测试获取整数配置
func TestGetInt_获取整数(t *testing.T) {
	c := New()
	c.Set("int.key", 42)

	value := c.GetInt("int.key")
	if value != 42 {
		t.Errorf("期望 42, 得到: %d", value)
	}
}

// TestGetInt64_获取64位整数 测试获取 64 位整数配置
func TestGetInt64_获取64位整数(t *testing.T) {
	c := New()
	c.Set("int64.key", int64(1234567890))

	value := c.GetInt64("int64.key")
	if value != int64(1234567890) {
		t.Errorf("期望 1234567890, 得到: %d", value)
	}
}

// TestGetBool_获取布尔值 测试获取布尔值配置
func TestGetBool_获取布尔值(t *testing.T) {
	c := New()
	c.Set("bool.key", true)

	value := c.GetBool("bool.key")
	if value != true {
		t.Errorf("期望 true, 得到: %v", value)
	}
}

// TestGetStringSlice_获取字符串切片 测试获取字符串切片配置
func TestGetStringSlice_获取字符串切片(t *testing.T) {
	c := New()
	c.Set("slice.key", []string{"a", "b", "c"})

	value := c.GetStringSlice("slice.key")
	if len(value) != 3 {
		t.Errorf("期望长度 3, 得到: %d", len(value))
	}
}

// TestGetStringMap_获取字符串Map 测试获取字符串 Map 配置
func TestGetStringMap_获取字符串Map(t *testing.T) {
	c := New()
	c.Set("map.key", map[string]interface{}{"key1": "value1"})

	value := c.GetStringMap("map.key")
	if value["key1"] != "value1" {
		t.Errorf("期望 value1, 得到: %v", value["key1"])
	}
}

// TestIsSet_检查键是否存在 测试检查键是否已设置
func TestIsSet_检查键是否存在(t *testing.T) {
	c := New()
	c.Set("exists.key", "value")

	if !c.IsSet("exists.key") {
		t.Error("键应该存在")
	}

	if c.IsSet("notexists.key") {
		t.Error("键不应该存在")
	}
}

// TestSet_设置值 测试设置配置值
func TestSet_设置值(t *testing.T) {
	c := New()
	c.Set("new.key", "new-value")

	value := c.Get("new.key")
	if value != "new-value" {
		t.Errorf("期望 new-value, 得到: %v", value)
	}
}

// TestSetDefault_设置默认值 测试设置默认值
func TestSetDefault_设置默认值(t *testing.T) {
	c := New()
	c.SetDefault("default.key", "default-value")

	value := c.Get("default.key")
	if value != "default-value" {
		t.Errorf("期望 default-value, 得到: %v", value)
	}
}

// TestUnmarshal_解marshal结构体 测试解 marshal 结构体
func TestUnmarshal_解marshal结构体(t *testing.T) {
	c := New()
	c.Set("server.port", 8080)
	c.Set("server.host", "localhost")

	type ServerConfig struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}

	var config ServerConfig
	err := c.Unmarshal(&config)
	if err != nil {
		t.Logf("解 marshal 错误: %v", err)
	}

	t.Logf("配置: %+v", config)
}

// TestUnmarshalKey_解marshal特定键 测试解 marshal 特定键
func TestUnmarshalKey_解marshal特定键(t *testing.T) {
	c := New()
	c.Set("database.host", "127.0.0.1")

	type DBConfig struct {
		Host string `mapstructure:"host"`
	}

	var config DBConfig
	err := c.UnmarshalKey("database", &config)
	if err != nil {
		t.Logf("解 marshal 错误: %v", err)
	}

	t.Logf("数据库配置: %+v", config)
}

// TestSub_获取子配置 测试获取子配置
func TestSub_获取子配置(t *testing.T) {
	c := New()
	c.Set("parent.child", "value")

	sub := c.Sub("parent")
	if sub == nil {
		t.Error("子配置获取失败")
	}
}

// TestAllSettings_获取所有设置 测试获取所有设置
func TestAllSettings_获取所有设置(t *testing.T) {
	c := New()
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	settings := c.AllSettings()
	if len(settings) == 0 {
		t.Error("设置获取失败")
	}

	t.Logf("所有设置: %+v", settings)
}

// TestGetViper_获取原始Viper 测试获取原始 Viper 实例
func TestGetViper_获取原始Viper(t *testing.T) {
	c := New()
	v := c.GetViper()

	if v == nil {
		t.Error("Viper 实例获取失败")
	}
}

// TestOptions_配置选项 测试配置选项
func TestOptions_配置选项(t *testing.T) {
	// 测试 WithConfigType
	c := New()
	err := c.Load("", WithConfigType("yaml"))
	if err != nil {
		t.Logf("加载错误: %v", err)
	}
}

// TestOptions_WithConfigName 测试配置名称选项
func TestOptions_WithConfigName(t *testing.T) {
	c := New()
	c.SetConfigName("custom-config")

	name := c.GetViper().ConfigFileUsed()
	t.Logf("配置文件: %s", name)
}

// TestOptions_WithSearchDirs 测试搜索目录选项
func TestOptions_WithSearchDirs(t *testing.T) {
	c := New()
	dirs := []string{"./", "./configs", "/etc"}
	_ = c.Load("", WithSearchDirs(dirs))
}

// TestOptions_WithDefaults 测试默认值选项
func TestOptions_WithDefaults(t *testing.T) {
	c := New()
	_ = c.Load("", WithDefault("key1", "value1"), WithDefaults(map[string]interface{}{
		"key2": "value2",
	}))

	if c.GetString("key1") != "value1" {
		t.Error("默认值设置失败")
	}
}

// TestLoadFromBytes_从字节加载 测试从字节数组加载配置
func TestLoadFromBytes_从字节加载(t *testing.T) {
	c := New()
	yamlContent := []byte("server:\n  port: 8080\n  host: localhost")

	err := c.LoadFromBytes(yamlContent, "yaml")
	if err != nil {
		t.Logf("加载错误: %v", err)
	}

	port := c.GetInt("server.port")
	t.Logf("端口: %d", port)
}

// TestGetSizeInBytes_获取字节大小 测试获取字节大小
func TestGetSizeInBytes_获取字节大小(t *testing.T) {
	c := New()
	c.Set("size", "10KB")

	size := c.GetSizeInBytes("size")
	t.Logf("大小: %s", size)
}

// TestMultipleConfigs_多个配置 测试多个配置
func TestMultipleConfigs_多个配置(t *testing.T) {
	configs := []*Config{
		New(),
		New(),
		New(),
	}

	for i, c := range configs {
		c.Set("key", i)
		if c.Get("key") != i {
			t.Errorf("配置 %d 设置失败", i)
		}
	}
}

// TestNestedKeys_嵌套键 测试嵌套键
func TestNestedKeys_嵌套键(t *testing.T) {
	c := New()
	c.Set("database.connection.host", "localhost")
	c.Set("database.connection.port", 5432)
	c.Set("database.pool.max", 10)

	host := c.GetString("database.connection.host")
	port := c.GetInt("database.connection.port")
	max := c.GetInt("database.pool.max")

	if host != "localhost" {
		t.Errorf("主机期望 localhost, 得到: %s", host)
	}
	if port != 5432 {
		t.Errorf("端口期望 5432, 得到: %d", port)
	}
	if max != 10 {
		t.Errorf("最大连接期望 10, 得到: %d", max)
	}
}

// TestEmptyConfig_空配置 测试空配置
func TestEmptyConfig_空配置(t *testing.T) {
	c := New()

	if c.Get("nonexistent") != nil {
		t.Error("不存在键应返回 nil")
	}

	if c.GetString("nonexistent") != "" {
		t.Error("不存在键应返回空字符串")
	}

	if c.GetInt("nonexistent") != 0 {
		t.Error("不存在键应返回 0")
	}
}

// TestConfigOverride_配置覆盖 测试配置覆盖
func TestConfigOverride_配置覆盖(t *testing.T) {
	c := New()

	c.Set("key", "value1")
	if c.GetString("key") != "value1" {
		t.Error("第一次设置失败")
	}

	c.Set("key", "value2")
	if c.GetString("key") != "value2" {
		t.Error("覆盖设置失败")
	}
}
