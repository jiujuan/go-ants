package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache 是缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (interface{}, error)
	// Set 设置缓存值
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// Clear 清空所有缓存
	Clear(ctx context.Context) error
}

// ===== 内存缓存实现 =====

// MemoryCache 内存缓存
type MemoryCache struct {
	data       map[string]*cacheItem
	mu         sync.RWMutex
	gcInterval time.Duration
	stopCh     chan struct{}
}

// cacheItem 缓存项
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache(opts ...MemoryOption) *MemoryCache {
	options := &MemoryOptions{
		gcInterval:      time.Minute,
		cleanupDisabled: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	cache := &MemoryCache{
		data:       make(map[string]*cacheItem),
		gcInterval: options.gcInterval,
		stopCh:     make(chan struct{}),
	}

	// 启动 GC 协程
	if !options.cleanupDisabled {
		go cache.gc()
	}

	return cache
}

// Get 获取缓存值
func (m *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	// 检查过期
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, ErrCacheMiss
	}

	return item.value, nil
}

// Set 设置缓存值
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	exp := time.Time{}
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.data[key] = &cacheItem{
		value:      value,
		expiration: exp,
	}

	return nil
}

// Delete 删除缓存
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Exists 检查键是否存在
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return false, nil
	}

	// 检查过期
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return false, nil
	}

	return true, nil
}

// Clear 清空所有缓存
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*cacheItem)
	return nil
}

// Stop 停止 GC
func (m *MemoryCache) Stop() {
	close(m.stopCh)
}

// gc 清理过期缓存项
func (m *MemoryCache) gc() {
	ticker := time.NewTicker(m.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if !item.expiration.IsZero() && now.After(item.expiration) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		case <-m.stopCh:
			return
		}
	}
}

// MemoryOption 内存缓存选项
type MemoryOption func(*MemoryOptions)

type MemoryOptions struct {
	gcInterval      time.Duration
	cleanupDisabled bool
}

// WithMemoryGCInterval 设置 GC 间隔
func WithMemoryGCInterval(interval time.Duration) MemoryOption {
	return func(o *MemoryOptions) {
		o.gcInterval = interval
	}
}

// WithMemoryCleanupDisabled 禁用自动清理
func WithMemoryCleanupDisabled(disabled bool) MemoryOption {
	return func(o *MemoryOptions) {
		o.cleanupDisabled = disabled
	}
}

// ===== Redis 缓存实现 =====

const (
	defaultRedisPrefix = "cache:"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client     *redis.Client
	prefix     string
	serializer Serializer
}

// Serializer 序列化器接口
type Serializer interface {
	Serialize(value interface{}) ([]byte, error)
	Deserialize(data []byte) (interface{}, error)
}

// JSONSerializer JSON 序列化器
type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (j *JSONSerializer) Deserialize(data []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(data, &result)
	return result, err
}

// NewRedisCache 创建新的 Redis 缓存
func NewRedisCache(client *redis.Client, opts ...RedisCacheOption) *RedisCache {
	options := &RedisCacheOptions{
		prefix:     defaultRedisPrefix,
		serializer: &JSONSerializer{},
	}

	for _, opt := range opts {
		opt(options)
	}

	return &RedisCache{
		client:     client,
		prefix:     options.prefix,
		serializer: options.serializer,
	}
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	key = r.prefix + key

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	return r.serializer.Deserialize(data)
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	key = r.prefix + key

	data, err := r.serializer.Serialize(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Delete 删除缓存
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	key = r.prefix + key
	return r.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	key = r.prefix + key

	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// Clear 清空所有缓存（使用 SCAN 遍历删除）
func (r *RedisCache) Clear(ctx context.Context) error {
	iter := r.client.Scan(ctx, 0, r.prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// RedisCacheOption Redis 缓存选项
type RedisCacheOption func(*RedisCacheOptions)

type RedisCacheOptions struct {
	prefix     string
	serializer Serializer
}

// WithRedisCachePrefix 设置键前缀
func WithRedisCachePrefix(prefix string) RedisCacheOption {
	return func(o *RedisCacheOptions) {
		o.prefix = prefix
	}
}

// WithRedisCacheSerializer 设置序列化器
func WithRedisCacheSerializer(serializer Serializer) RedisCacheOption {
	return func(o *RedisCacheOptions) {
		o.serializer = serializer
	}
}

// ===== 错误定义 =====

var (
	ErrCacheMiss         = fmt.Errorf("cache miss")
	ErrInvalidExpiration = fmt.Errorf("invalid expiration")
)

// ===== 缓存工具函数 =====

// Cacher 是带泛型的缓存接口
type Cacher[T any] interface {
	Get(ctx context.Context, key string) (T, error)
	Set(ctx context.Context, key string, value T, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// GenericCache 泛型缓存封装
type GenericCache[T any] struct {
	cache Cache
}

// NewGenericCache 创建泛型缓存
func NewGenericCache[T any](cache Cache) *GenericCache[T] {
	return &GenericCache[T]{cache: cache}
}

// Get 获取缓存值（泛型）
func (g *GenericCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T
	val, err := g.cache.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	if v, ok := val.(T); ok {
		return v, nil
	}

	return zero, fmt.Errorf("type mismatch")
}

// Set 设置缓存值（泛型）
func (g *GenericCache[T]) Set(ctx context.Context, key string, value T, expiration time.Duration) error {
	return g.cache.Set(ctx, key, value, expiration)
}

// Delete 删除缓存
func (g *GenericCache[T]) Delete(ctx context.Context, key string) error {
	return g.cache.Delete(ctx, key)
}

// ===== 缓存包装器（支持缓存空值） =====

// nullMarker 用于标记空值
var nullMarker = struct{}{}

// NullableCache 支持缓存空值的包装器
type NullableCache struct {
	cache   Cache
	nullTTL time.Duration
}

// NewNullableCache 创建支持空值的缓存
func NewNullableCache(cache Cache, nullTTL time.Duration) *NullableCache {
	return &NullableCache{
		cache:   cache,
		nullTTL: nullTTL,
	}
}

// Get 获取缓存值
func (n *NullableCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := n.cache.Get(ctx, key)
	if err == ErrCacheMiss {
		return nil, ErrCacheMiss
	}

	if val == nullMarker {
		return nil, nil
	}

	return val, err
}

// Set 设置缓存值
func (n *NullableCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if value == nil {
		return n.cache.Set(ctx, key, nullMarker, n.nullTTL)
	}
	return n.cache.Set(ctx, key, value, expiration)
}

// Delete 删除缓存
func (n *NullableCache) Delete(ctx context.Context, key string) error {
	return n.cache.Delete(ctx, key)
}

// Exists 检查键是否存在
func (n *NullableCache) Exists(ctx context.Context, key string) (bool, error) {
	return n.cache.Exists(ctx, key)
}

// Clear 清空所有缓存
func (n *NullableCache) Clear(ctx context.Context) error {
	return n.cache.Clear(ctx)
}

// ===== 缓存装饰器 =====

// CacheWithMetrics 带指标的缓存装饰器
type CacheWithMetrics struct {
	cache  Cache
	hits   int64
	misses int64
	mu     sync.RWMutex
}

// NewCacheWithMetrics 创建带指标的缓存
func NewCacheWithMetrics(cache Cache) *CacheWithMetrics {
	return &CacheWithMetrics{cache: cache}
}

// Get 获取缓存值
func (m *CacheWithMetrics) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := m.cache.Get(ctx, key)
	if err == nil {
		m.mu.Lock()
		m.hits++
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		m.misses++
		m.mu.Unlock()
	}
	return val, err
}

// Set 设置缓存值
func (m *CacheWithMetrics) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.cache.Set(ctx, key, value, expiration)
}

// Delete 删除缓存
func (m *CacheWithMetrics) Delete(ctx context.Context, key string) error {
	return m.cache.Delete(ctx, key)
}

// Exists 检查键是否存在
func (m *CacheWithMetrics) Exists(ctx context.Context, key string) (bool, error) {
	return m.cache.Exists(ctx, key)
}

// Clear 清空所有缓存
func (m *CacheWithMetrics) Clear(ctx context.Context) error {
	m.mu.Lock()
	m.hits = 0
	m.misses = 0
	m.mu.Unlock()
	return m.cache.Clear(ctx)
}

// Stats 获取缓存统计信息
func (m *CacheWithMetrics) Stats() (hits, misses int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hits, m.misses
}
