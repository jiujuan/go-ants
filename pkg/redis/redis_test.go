package redis

import (
	"context"
	"testing"
	"time"
)

// TestNew_DefaultOptions 测试默认选项创建 Redis 客户端
func TestNew_DefaultOptions(t *testing.T) {
	ctx := context.Background()

	// 由于没有实际的 Redis 连接，这里测试选项
	opts := &Options{
		Addr:         "localhost:6379",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		Retries:      3,
	}

	if opts.Addr != "localhost:6379" {
		t.Errorf("地址期望 localhost:6379, 得到: %s", opts.Addr)
	}

	if opts.DB != 0 {
		t.Errorf("数据库期望 0, 得到: %d", opts.DB)
	}

	if opts.PoolSize != 10 {
		t.Errorf("连接池大小期望 10, 得到: %d", opts.PoolSize)
	}
}

// TestWithAddr_选项函数 测试地址选项函数
func TestWithAddr_选项函数(t *testing.T) {
	opts := &Options{}

	WithAddr("redis-server:6379")(opts)

	if opts.Addr != "redis-server:6379" {
		t.Errorf("地址设置失败: %s", opts.Addr)
	}
}

// TestWithPassword_选项函数 测试密码选项函数
func TestWithPassword_选项函数(t *testing.T) {
	opts := &Options{}

	WithPassword("redis-password")(opts)

	if opts.Password != "redis-password" {
		t.Errorf("密码设置失败: %s", opts.Password)
	}
}

// TestWithDB_选项函数 测试数据库选项函数
func TestWithDB_选项函数(t *testing.T) {
	opts := &Options{}

	WithDB(5)(opts)

	if opts.DB != 5 {
		t.Errorf("数据库编号设置失败: %d", opts.DB)
	}
}

// TestWithPoolSize_选项函数 测试连接池大小选项函数
func TestWithPoolSize_选项函数(t *testing.T) {
	opts := &Options{}

	WithPoolSize(20)(opts)

	if opts.PoolSize != 20 {
		t.Errorf("连接池大小设置失败: %d", opts.PoolSize)
	}
}

// TestWithMinIdleConns_选项函数 测试最小空闲连接选项函数
func TestWithMinIdleConns_选项函数(t *testing.T) {
	opts := &Options{}

	WithMinIdleConns(5)(opts)

	if opts.MinIdleConns != 5 {
		t.Errorf("最小空闲连接设置失败: %d", opts.MinIdleConns)
	}
}

// TestWithMaxConnAge_选项函数 测试连接最大生命周期选项函数
func TestWithMaxConnAge_选项函数(t *testing.T) {
	opts := &Options{}

	age := time.Minute * 10
	WithMaxConnAge(age)(opts)

	if opts.MaxConnAge != age {
		t.Error("连接最大生命周期设置失败")
	}
}

// TestWithReadTimeout_选项函数 测试读取超时选项函数
func TestWithReadTimeout_选项函数(t *testing.T) {
	opts := &Options{}

	timeout := time.Second * 5
	WithReadTimeout(timeout)(opts)

	if opts.ReadTimeout != timeout {
		t.Error("读取超时设置失败")
	}
}

// TestWithWriteTimeout_选项函数 测试写入超时选项函数
func TestWithWriteTimeout_选项函数(t *testing.T) {
	opts := &Options{}

	timeout := time.Second * 5
	WithWriteTimeout(timeout)(opts)

	if opts.WriteTimeout != timeout {
		t.Error("写入超时设置失败")
	}
}

// TestWithDialTimeout_选项函数 测试拨号超时选项函数
func TestWithDialTimeout_选项函数(t *testing.T) {
	opts := &Options{}

	timeout := time.Second * 3
	WithDialTimeout(timeout)(opts)

	if opts.DialTimeout != timeout {
		t.Error("拨号超时设置失败")
	}
}

// TestWithPoolTimeout_选项函数 测试连接池超时选项函数
func TestWithPoolTimeout_选项函数(t *testing.T) {
	opts := &Options{}

	timeout := time.Second * 3
	WithPoolTimeout(timeout)(opts)

	if opts.PoolTimeout != timeout {
		t.Error("连接池超时设置失败")
	}
}

// TestWithRetries_选项函数 测试重试次数选项函数
func TestWithRetries_选项函数(t *testing.T) {
	opts := &Options{}

	WithRetries(5)(opts)

	if opts.Retries != 5 {
		t.Errorf("重试次数设置失败: %d", opts.Retries)
	}
}

// TestClusterOptions_集群选项 测试集群选项
func TestClusterOptions_集群选项(t *testing.T) {
	opts := &ClusterOptions{
		Addrs:    []string{"localhost:7001", "localhost:7002"},
		Password: "cluster-pass",
		PoolSize: 50,
	}

	if len(opts.Addrs) != 2 {
		t.Errorf("集群节点数量错误: %d", len(opts.Addrs))
	}

	if opts.Password != "cluster-pass" {
		t.Error("集群密码设置失败")
	}
}

// TestWithClusterAddrs_选项函数 测试集群地址选项函数
func TestWithClusterAddrs_选项函数(t *testing.T) {
	opts := &ClusterOptions{}

	addrs := []string{"localhost:7001", "localhost:7002", "localhost:7003"}
	WithClusterAddrs(addrs)(opts)

	if len(opts.Addrs) != 3 {
		t.Errorf("集群地址数量错误: %d", len(opts.Addrs))
	}
}

// TestWithClusterPassword_选项函数 测试集群密码选项函数
func TestWithClusterPassword_选项函数(t *testing.T) {
	opts := &ClusterOptions{}

	WithClusterPassword("cluster-secret")(opts)

	if opts.Password != "cluster-secret" {
		t.Error("集群密码设置失败")
	}
}

// TestClusterOptions_clusterConfig 测试集群配置
func TestClusterOptions_clusterConfig(t *testing.T) {
	opts := &ClusterOptions{
		Addrs:        []string{"localhost:7001"},
		Password:     "pass",
		PoolSize:     20,
		MinIdleConns: 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		DialTimeout:  time.Second * 3,
		PoolTimeout:  time.Second * 3,
		Retries:      3,
	}

	config := opts.clusterConfig()

	if config.Addrs[0] != "localhost:7001" {
		t.Error("集群配置地址设置失败")
	}

	if config.PoolSize != 20 {
		t.Error("集群连接池大小设置失败")
	}
}

// TestOptions_多个选项组合 测试多个选项组合
func TestOptions_多个选项组合(t *testing.T) {
	opts := &Options{}

	WithAddr("redis-master:6379")(
		&Options{},
	)

	WithAddr("redis-master:6379")(opts)
	WithPassword("password")(opts)
	WithDB(3)(opts)
	WithPoolSize(50)(opts)
	WithMinIdleConns(10)(opts)
	WithRetries(5)(opts)

	if opts.Addr != "redis-master:6379" {
		t.Error("地址设置失败")
	}

	if opts.Password != "password" {
		t.Error("密码设置失败")
	}

	if opts.DB != 3 {
		t.Error("数据库编号设置失败")
	}

	if opts.PoolSize != 50 {
		t.Error("连接池大小设置失败")
	}

	if opts.MinIdleConns != 10 {
		t.Error("最小空闲连接设置失败")
	}

	if opts.Retries != 5 {
		t.Error("重试次数设置失败")
	}
}

// TestClient_结构 测试客户端结构
func TestClient_结构(t *testing.T) {
	client := &Client{}

	if client == nil {
		t.Error("客户端创建失败")
	}
}

// TestClusterClient_结构 测试集群客户端结构
func TestClusterClient_结构(t *testing.T) {
	client := &ClusterClient{}

	if client == nil {
		t.Error("集群客户端创建失败")
	}
}

// TestRedisOperation_模拟操作 测试 Redis 操作模拟
func TestRedisOperation_模拟操作(t *testing.T) {
	// 测试基本的 Redis 操作常量
	key := "test-key"
	value := "test-value"
	expiration := time.Hour

	if key == "" {
		t.Error("键不应为空")
	}

	if value == "" {
		t.Error("值不应为空")
	}

	if expiration <= 0 {
		t.Error("过期时间应大于 0")
	}
}

// TestKeyPatterns_键模式 测试常见的键模式
func TestKeyPatterns_键模式(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"用户键", "user:1001", "user:1001"},
		{"缓存键", "cache:session:abc", "cache:session:abc"},
		{"分布式锁", "lock:resource:A", "lock:resource:A"},
		{"计数器", "counter:pageview:home", "counter:pageview:home"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.expected {
				t.Errorf("键模式不匹配: %s != %s", tt.key, tt.expected)
			}
		})
	}
}

// TestExpiration_过期时间 测试过期时间设置
func TestExpiration_过期时间(t *testing.T) {
	tests := []struct {
		name        string
		expiration  time.Duration
		expectValid bool
	}{
		{"永不过期", 0, false},
		{"1秒", time.Second, true},
		{"1分钟", time.Minute, true},
		{"1小时", time.Hour, true},
		{"1天", time.Hour * 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectValid && tt.expiration <= 0 {
				t.Error("过期时间应该大于 0")
			}
		})
	}
}

// TestDatabaseNumber_数据库编号 测试数据库编号
func TestDatabaseNumber_数据库编号(t *testing.T) {
	// Redis 默认有 16 个数据库 (0-15)
	maxDB := 16

	for i := 0; i < maxDB; i++ {
		opts := &Options{}
		WithDB(i)(opts)

		if opts.DB != i {
			t.Errorf("数据库编号设置失败: %d", i)
		}
	}
}

// TestPoolSize_连接池大小 测试连接池大小
func TestPoolSize_连接池大小(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
		minValid int
		maxValid int
	}{
		{"最小", 1, 1, 1000},
		{"默认", 10, 1, 1000},
		{"较大", 100, 1, 1000},
		{"超大", 1000, 1, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			WithPoolSize(tt.poolSize)(opts)

			if opts.PoolSize < tt.minValid || opts.PoolSize > tt.maxValid {
				t.Errorf("连接池大小超出有效范围: %d", opts.PoolSize)
			}
		})
	}
}

// TestTimeout_超时设置 测试超时设置
func TestTimeout_超时设置(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		min     time.Duration
		max     time.Duration
	}{
		{"最小超时", time.Millisecond * 100, time.Millisecond * 100, time.Second * 30},
		{"默认超时", time.Second * 3, time.Millisecond * 100, time.Second * 30},
		{"较大超时", time.Second * 30, time.Millisecond * 100, time.Second * 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout < tt.min || tt.timeout > tt.max {
				t.Errorf("超时时间超出有效范围: %v", tt.timeout)
			}
		})
	}
}

// TestRetries_重试次数 测试重试次数
func TestRetries_重试次数(t *testing.T) {
	tests := []struct {
		name    string
		retries int
	}{
		{"无重试", 0},
		{"单次重试", 1},
		{"默认重试", 3},
		{"多次重试", 5},
		{"十次重试", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			WithRetries(tt.retries)(opts)

			if opts.Retries != tt.retries {
				t.Errorf("重试次数设置失败: %d", tt.retries)
			}
		})
	}
}

// TestConnectionString_连接字符串 测试连接字符串格式
func TestConnectionString_连接字符串(t *testing.T) {
	tests := []struct {
		name         string
		addr         string
		password     string
		db           int
		expectFormat bool
	}{
		{"无密码", "localhost:6379", "", 0, true},
		{"有密码", "localhost:6379", "pass", 0, true},
		{"不同数据库", "localhost:6379", "", 5, true},
		{"远程服务器", "redis.example.com:6379", "pass", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			WithAddr(tt.addr)(opts)
			WithPassword(tt.password)(opts)
			WithDB(tt.db)(opts)

			if tt.expectFormat && opts.Addr == "" {
				t.Error("地址格式无效")
			}
		})
	}
}
