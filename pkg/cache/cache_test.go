package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_Basic(t *testing.T) {
	cache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer cache.Stop()

	ctx := context.Background()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	val, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Get non-existent
	_, err = cache.Get(ctx, "nonexistent")
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer cache.Stop()

	ctx := context.Background()

	// Set with short expiration
	err := cache.Set(ctx, "expiry", "value", 50*time.Millisecond)
	assert.NoError(t, err)

	// Should exist immediately
	_, err = cache.Get(ctx, "expiry")
	assert.NoError(t, err)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "expiry")
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer cache.Stop()

	ctx := context.Background()

	err := cache.Set(ctx, "delete_key", "value", 0)
	assert.NoError(t, err)

	err = cache.Delete(ctx, "delete_key")
	assert.NoError(t, err)

	_, err = cache.Get(ctx, "delete_key")
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer cache.Stop()

	ctx := context.Background()

	err := cache.Set(ctx, "exists", "value", 0)
	assert.NoError(t, err)

	exists, err := cache.Exists(ctx, "exists")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Exists(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer cache.Stop()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 0)
	cache.Set(ctx, "key2", "value2", 0)

	err := cache.Clear(ctx)
	assert.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	assert.Equal(t, ErrCacheMiss, err)
}

func TestJSONSerializer(t *testing.T) {
	serializer := &JSONSerializer{}

	data, err := serializer.Serialize("test")
	assert.NoError(t, err)

	val, err := serializer.Deserialize(data)
	assert.NoError(t, err)
	assert.Equal(t, "test", val)
}

func TestGenericCache(t *testing.T) {
	memoryCache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer memoryCache.Stop()

	genericCache := NewGenericCache[int](memoryCache)

	ctx := context.Background()

	err := genericCache.Set(ctx, "num", 42, 0)
	assert.NoError(t, err)

	val, err := genericCache.Get(ctx, "num")
	assert.NoError(t, err)
	assert.Equal(t, 42, val)
}

func TestNullableCache(t *testing.T) {
	memoryCache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer memoryCache.Stop()

	nullableCache := NewNullableCache(memoryCache, time.Minute)

	ctx := context.Background()

	err := nullableCache.Set(ctx, "nil_val", nil, 0)
	assert.NoError(t, err)

	val, err := nullableCache.Get(ctx, "nil_val")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestCacheWithMetrics(t *testing.T) {
	memoryCache := NewMemoryCache(WithMemoryCleanupDisabled(true))
	defer memoryCache.Stop()

	metricsCache := NewCacheWithMetrics(memoryCache)

	ctx := context.Background()

	metricsCache.Set(ctx, "key1", "value1", 0)
	metricsCache.Get(ctx, "key1")
	metricsCache.Get(ctx, "nonexistent")

	hits, misses := metricsCache.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)
}

func TestErrors(t *testing.T) {
	assert.Equal(t, "cache miss", ErrCacheMiss.Error())
	assert.Equal(t, "invalid expiration", ErrInvalidExpiration.Error())
}
