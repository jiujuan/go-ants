package wire

import (
	"github.com/google/wire"

	"github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/auth"
	"github.com/jiujuan/go-ants/pkg/cache"
	"github.com/jiujuan/go-ants/pkg/conf"
	"github.com/jiujuan/go-ants/pkg/database"
	"github.com/jiujuan/go-ants/pkg/es"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/metric"
	"github.com/jiujuan/go-ants/pkg/mq"
	"github.com/jiujuan/go-ants/pkg/redis"
	"github.com/jiujuan/go-ants/pkg/transport"
	"github.com/jiujuan/go-ants/pkg/validator"
	"github.com/jiujuan/go-ants/pkg/worker"
)

// ===== 应用层依赖注入 =====

// AppSet 应用组件集合
var AppSet = wire.NewSet(
	app.New,
	app.WithName,
	app.WithLogger,
)

// LoggerSet 日志依赖集合
var LoggerSet = wire.NewSet(
	log.New,
	log.DefaultLogger,
)

// ConfigSet 配置依赖集合
var ConfigSet = wire.NewSet(
	conf.New,
)

// ValidatorSet 验证器依赖集合
var ValidatorSet = wire.NewSet(
	validator.New,
	validator.Default,
)

// DatabaseSet 数据库依赖集合
var DatabaseSet = wire.NewSet(
	database.New,
)

// RedisSet Redis 依赖集合
var RedisSet = wire.NewSet(
	redis.New,
)

// CacheSet 缓存依赖集合
var CacheSet = wire.NewSet(
	cache.NewMemoryCache,
	cache.NewRedisCache,
)

// AuthSet 认证依赖集合
var AuthSet = wire.NewSet(
	auth.New,
)

// MQSet 消息队列依赖集合
var MQSet = wire.NewSet(
	mq.NewKafkaProducer,
	mq.NewKafkaConsumer,
	mq.NewRabbitMQProducer,
	mq.NewRabbitMQConsumer,
)

// WorkerSet 工作池依赖集合
var WorkerSet = wire.NewSet(
	worker.New,
)

// ESSet Elasticsearch 依赖集合
var ESSet = wire.NewSet(
	es.New,
)

// TransportSet 传输层依赖集合
var TransportSet = wire.NewSet(
	transport.NewGinServer,
	transport.NewFiberServer,
	transport.NewMuxServer,
)

// MetricSet 监控依赖集合
var MetricSet = wire.NewSet(
	metric.InitMetrics,
	metric.StartPProf,
)

// ===== Provider 函数 =====

// Provider 是依赖提供者函数类型
type Provider func() (interface{}, error)

// Providers 依赖提供者映射
var Providers = make(map[string]Provider)

// RegisterProvider 注册依赖提供者
func RegisterProvider(name string, provider Provider) {
	Providers[name] = provider
}

// GetProvider 获取依赖提供者
func GetProvider(name string) (Provider, bool) {
	p, ok := Providers[name]
	return p, ok
}

// ===== Wire 生成标记 =====

//go:generate wire
