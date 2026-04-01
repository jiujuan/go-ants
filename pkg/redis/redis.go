package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/jiujuan/go-ants/pkg/log"
)

// Client Redis 客户端包装器
type Client struct {
	client *redis.Client
}

// Option 是 Redis 选项函数
type Option func(*Options)

// Options Redis 配置选项
type Options struct {
	// Addr Redis 地址
	Addr string
	// Password Redis 密码
	Password string
	// DB Redis 数据库编号
	DB int
	// PoolSize 连接池大小
	PoolSize int
	// MinIdleConns 最小空闲连接数
	MinIdleConns int
	// MaxConnAge 连接最大生命周期
	MaxConnAge time.Duration
	// ReadTimeout 读取超时
	ReadTimeout time.Duration
	// WriteTimeout 写入超时
	WriteTimeout time.Duration
	// DialTimeout 拨号超时
	DialTimeout time.Duration
	// PoolTimeout 连接池超时
	PoolTimeout time.Duration
	// Retries 重试次数
	Retries int
	// Ranger 范围操作
	Ranger *redis.Ranger
	// TLSConfig TLS 配置
	TLSConfig *redis.TLSConfig
}

// WithAddr 设置 Redis 地址
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithPassword 设置 Redis 密码
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithDB 设置数据库编号
func WithDB(db int) Option {
	return func(o *Options) {
		o.DB = db
	}
}

// WithPoolSize 设置连接池大小
func WithPoolSize(poolSize int) Option {
	return func(o *Options) {
		o.PoolSize = poolSize
	}
}

// WithMinIdleConns 设置最小空闲连接数
func WithMinIdleConns(minIdleConns int) Option {
	return func(o *Options) {
		o.MinIdleConns = minIdleConns
	}
}

// WithMaxConnAge 设置连接最大生命周期
func WithMaxConnAge(maxConnAge time.Duration) Option {
	return func(o *Options) {
		o.MaxConnAge = maxConnAge
	}
}

// WithReadTimeout 设置读取超时
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout 设置写入超时
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = writeTimeout
	}
}

// WithDialTimeout 设置拨号超时
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = dialTimeout
	}
}

// WithPoolTimeout 设置连接池超时
func WithPoolTimeout(poolTimeout time.Duration) Option {
	return func(o *Options) {
		o.PoolTimeout = poolTimeout
	}
}

// WithRetries 设置重试次数
func WithRetries(retries int) Option {
	return func(o *Options) {
		o.Retries = retries
	}
}

// WithTLSConfig 设置 TLS 配置
func WithTLSConfig(tlsConfig *redis.TLSConfig) Option {
	return func(o *Options) {
		o.TLSConfig = tlsConfig
	}
}

// New 创建新的 Redis 客户端
func New(ctx context.Context, opts ...Option) (*Client, error) {
	options := &Options{
		Addr:         "localhost:6379",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxConnAge:   time.Second * 30,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		DialTimeout:  time.Second * 3,
		PoolTimeout:  time.Second * 3,
		Retries:      3,
	}

	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         options.Addr,
		Password:     options.Password,
		DB:           options.DB,
		PoolSize:     options.PoolSize,
		MinIdleConns: options.MinIdleConns,
		MaxConnAge:   options.MaxConnAge,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		DialTimeout:  options.DialTimeout,
		PoolTimeout:  options.PoolTimeout,
		Retries:      options.Retries,
		TLSConfig:    options.TLSConfig,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("redis connected",
		log.String("addr", options.Addr),
		log.Int("db", options.DB))

	return &Client{client: client}, nil
}

// NewCluster 创建 Redis 集群客户端
func NewCluster(ctx context.Context, opts ...ClusterOption) (*ClusterClient, error) {
	options := &ClusterOptions{
		ClusterSlots: redis.ClusterSlots,
		PoolSize:     100,
	}

	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClusterClient(options.clusterConfig())

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis cluster: %w", err)
	}

	log.Info("redis cluster connected")

	return &ClusterClient{client: client}, nil
}

// ClusterOptions 集群选项
type ClusterOptions struct {
	Addrs        []string
	Password     string
	PoolSize     int
	MinIdleConns int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
	PoolTimeout  time.Duration
	Retries      int
}

// ClusterOption 集群选项函数
type ClusterOption func(*ClusterOptions)

// WithClusterAddrs 设置集群节点地址
func WithClusterAddrs(addrs []string) ClusterOption {
	return func(o *ClusterOptions) {
		o.Addrs = addrs
	}
}

// WithClusterPassword 设置集群密码
func WithClusterPassword(password string) ClusterOption {
	return func(o *ClusterOptions) {
		o.Password = password
	}
}

func (o *ClusterOptions) clusterConfig() *redis.ClusterConfig {
	return &redis.ClusterConfig{
		Addrs:        o.Addrs,
		Password:     o.Password,
		PoolSize:     o.PoolSize,
		MinIdleConns: o.MinIdleConns,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
		DialTimeout:  o.DialTimeout,
		PoolTimeout:  o.PoolTimeout,
		Retries:      o.Retries,
	}
}

// ClusterClient Redis 集群客户端
type ClusterClient struct {
	client *redis.ClusterClient
}

// Client 获取底层客户端
func (c *Client) Client() *redis.Client {
	return c.client
}

// Cluster 获取底层集群客户端
func (c *ClusterClient) Cluster() *redis.ClusterClient {
	return c.client
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.client.Close()
}

// Close 关闭集群客户端
func (c *ClusterClient) Close() error {
	return c.client.Close()
}

// ===== 基础操作 =====

// Set 设置键值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// TTL 获取剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Persist 移除过期时间
func (c *Client) Persist(ctx context.Context, key string) (bool, error) {
	return c.client.Persist(ctx, key).Result()
}

// ===== 字符串操作 =====

// SetNX 设置键值（仅当键不存在）
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// MGet 批量获取
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.client.MGet(ctx, keys...).Result()
}

// MSet 批量设置
func (c *Client) MSet(ctx context.Context, values ...interface{}) error {
	return c.client.MSet(ctx, values...).Err()
}

// Incr 递增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy 增加指定值
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr 递减
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy 减少指定值
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// ===== 哈希操作 =====

// HSet 设置哈希字段
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.client.HDel(ctx, key, fields...).Result()
}

// HExists 检查哈希字段是否存在
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// HIncrBy 增加哈希字段值
func (c *Client) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return c.client.HIncrBy(ctx, key, field, incr).Result()
}

// HLen 获取哈希字段数量
func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	return c.client.HLen(ctx, key).Result()
}

// ===== 列表操作 =====

// LPush 左侧入队
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.LPush(ctx, key, values...).Result()
}

// RPush 右侧入队
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.RPush(ctx, key, values...).Result()
}

// LPop 左侧出队
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// RPop 右侧出队
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LRange 获取列表范围
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// ===== 集合操作 =====

// SAdd 添加集合成员
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SAdd(ctx, key, members...).Result()
}

// SRem 移除集合成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SRem(ctx, key, members...).Result()
}

// SMembers 获取所有集合成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否是集合成员
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合基数
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}

// ===== 有序集合操作 =====

// ZAdd 添加有序集合成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return c.client.ZAdd(ctx, key, members...).Result()
}

// ZRem 移除有序集合成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.ZRem(ctx, key, members...).Result()
}

// ZRangeByScore 按分数范围获取
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRank 获取成员排名
func (c *Client) ZRank(ctx context.Context, key, member string) (int64, error) {
	return c.client.ZRank(ctx, key, member).Result()
}

// ZScore 获取成员分数
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.client.ZScore(ctx, key, member).Result()
}

// ===== 过期操作 =====

// ExpireAt 设置过期时间点
func (c *Client) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	return c.client.ExpireAt(ctx, key, tm).Result()
}

// ===== 管道操作 =====

// Pipeline 创建管道
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// TxPipeline 创建事务管道
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// ===== 事务操作 =====

// Watch 监视键
func (c *Client) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return c.client.Watch(ctx, fn, keys...)
}

// ===== 发布/订阅操作 =====

// Subscribe 订阅
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// Publish 发布
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.client.Publish(ctx, channel, message).Result()
}

// ===== 脚本操作 =====

// ScriptLoad 加载脚本
func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.client.ScriptLoad(ctx, script).Result()
}

// Eval 执行脚本
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}
