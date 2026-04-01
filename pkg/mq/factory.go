package mq

import (
	"fmt"
)

// ===== MQ 工厂管理器 =====

// FactoryManager MQ 工厂管理器
type FactoryManager struct {
	factories map[MQType]MQFactory
}

// NewFactoryManager 创建工厂管理器
func NewFactoryManager() *FactoryManager {
	return &FactoryManager{
		factories: make(map[MQType]MQFactory),
	}
}

// Register 注册 MQ 工厂
func (fm *FactoryManager) Register(factory MQFactory) {
	fm.factories[factory.GetMQType()] = factory
}

// GetFactory 获取指定类型的工厂
func (fm *FactoryManager) GetFactory(mqType MQType) (MQFactory, error) {
	factory, ok := fm.factories[mqType]
	if !ok {
		return nil, fmt.Errorf("mq factory not registered: %w", ErrMQNotSupported)
	}
	return factory, nil
}

// CreateProducer 创建生产者
func (fm *FactoryManager) CreateProducer(mqType MQType, opts ...ProducerOption) (Producer, error) {
	factory, err := fm.GetFactory(mqType)
	if err != nil {
		return nil, err
	}
	return factory.CreateProducer(opts...)
}

// CreateConsumer 创建消费者
func (fm *FactoryManager) CreateConsumer(mqType MQType, opts ...ConsumerOption) (Consumer, error) {
	factory, err := fm.GetFactory(mqType)
	if err != nil {
		return nil, err
	}
	return factory.CreateConsumer(opts...)
}

// ===== 全局工厂管理器实例 =====

var defaultManager *FactoryManager

// InitMQ 初始化 MQ 工厂管理器
func InitMQ() *FactoryManager {
	defaultManager = NewFactoryManager()

	// 注册默认的工厂
	defaultManager.Register(NewKafkaFactory())
	defaultManager.Register(NewRabbitMQFactory())

	return defaultManager
}

// GetMQManager 获取默认的 MQ 工厂管理器
func GetMQManager() *FactoryManager {
	if defaultManager == nil {
		return InitMQ()
	}
	return defaultManager
}

// ===== 便捷构造函数 =====

// NewProducer 创建一个生产者
// 使用示例:
//
//	producer, err := mq.NewProducer(mq.MQTypeKafka,
//	    mq.WithKafkaBrokers("localhost:9092"),
//	    mq.WithKafkaTopic("my-topic"))
func NewProducer(mqType MQType, opts ...ProducerOption) (Producer, error) {
	manager := GetMQManager()
	return manager.CreateProducer(mqType, opts...)
}

// NewConsumer 创建一个消费者
// 使用示例:
//
//	consumer, err := mq.NewConsumer(mq.MQTypeKafka,
//	    mq.WithKafkaConsumerBrokers("localhost:9092"),
//	    mq.WithKafkaConsumerTopic("my-topic"),
//	    mq.WithKafkaConsumerGroupID("my-group"))
func NewConsumer(mqType MQType, opts ...ConsumerOption) (Consumer, error) {
	manager := GetMQManager()
	return manager.CreateConsumer(mqType, opts...)
}

// ===== 统一的选项类型（支持不同 MQ） =====

// CommonProducerOptions 通用的生产者选项
type CommonProducerOptions struct {
	MQType   MQType
	Brokers  []string // Kafka brokers 或 RabbitMQ URL
	Topic    string   // Kafka topic 或 RabbitMQ exchange
	Username string   // RabbitMQ 用户名
	Password string   // RabbitMQ 密码
	VHost    string   // RabbitMQ 虚拟主机
}

// CommonProducerOption 通用生产者选项函数
type CommonProducerOption func(*CommonProducerOptions)

// WithCommonBrokers 设置 broker 列表或 URL
func WithCommonBrokers(brokers ...string) CommonProducerOption {
	return func(o *CommonProducerOptions) {
		o.Brokers = brokers
	}
}

// WithCommonTopic 设置主题
func WithCommonTopic(topic string) CommonProducerOption {
	return func(o *CommonProducerOptions) {
		o.Topic = topic
	}
}

// WithCommonMQType 设置 MQ 类型
func WithCommonMQType(mqType MQType) CommonProducerOption {
	return func(o *CommonProducerOptions) {
		o.MQType = mqType
	}
}

// CreateProducerWithCommonOpts 使用通用选项创建生产者
func CreateProducerWithCommonOpts(opts ...CommonProducerOption) (Producer, error) {
	options := &CommonProducerOptions{
		MQType:  MQTypeKafka,
		Brokers: []string{"localhost:9092"},
		Topic:   "default",
	}

	for _, opt := range opts {
		opt(options)
	}

	switch options.MQType {
	case MQTypeKafka:
		return NewKafkaProducer(
			WithKafkaBrokers(options.Brokers...),
			WithKafkaTopic(options.Topic),
		)
	case MQTypeRabbitMQ:
		url := "amqp://guest:guest@localhost:5672/"
		if len(options.Brokers) > 0 {
			url = options.Brokers[0]
		}
		return NewRabbitMQProducer(
			WithRabbitMQURL(url),
			WithRabbitMQExchange(options.Topic),
		)
	default:
		return nil, ErrMQNotSupported
	}
}

// CommonConsumerOptions 通用的消费者选项
type CommonConsumerOptions struct {
	MQType   MQType
	Brokers  []string
	Topic    string
	GroupID  string
	Username string
	Password string
	VHost    string
}

// CommonConsumerOption 通用消费者选项函数
type CommonConsumerOption func(*CommonConsumerOptions)

// WithCommonConsumerBrokers 设置 broker 列表或 URL
func WithCommonConsumerBrokers(brokers ...string) CommonConsumerOption {
	return func(o *CommonConsumerOptions) {
		o.Brokers = brokers
	}
}

// WithCommonConsumerTopic 设置主题
func WithCommonConsumerTopic(topic string) CommonConsumerOption {
	return func(o *CommonConsumerOptions) {
		o.Topic = topic
	}
}

// WithCommonConsumerGroupID 设置消费者组 ID
func WithCommonConsumerGroupID(groupID string) CommonConsumerOption {
	return func(o *CommonConsumerOptions) {
		o.GroupID = groupID
	}
}

// WithCommonConsumerMQType 设置 MQ 类型
func WithCommonConsumerMQType(mqType MQType) CommonConsumerOption {
	return func(o *CommonConsumerOptions) {
		o.MQType = mqType
	}
}

// CreateConsumerWithCommonOpts 使用通用选项创建消费者
func CreateConsumerWithCommonOpts(opts ...CommonConsumerOption) (Consumer, error) {
	options := &CommonConsumerOptions{
		MQType:  MQTypeKafka,
		Brokers: []string{"localhost:9092"},
		Topic:   "default",
		GroupID: "default-group",
	}

	for _, opt := range opts {
		opt(options)
	}

	switch options.MQType {
	case MQTypeKafka:
		return NewKafkaConsumer(
			WithKafkaConsumerBrokers(options.Brokers...),
			WithKafkaConsumerTopic(options.Topic),
			WithKafkaConsumerGroupID(options.GroupID),
		)
	case MQTypeRabbitMQ:
		url := "amqp://guest:guest@localhost:5672/"
		if len(options.Brokers) > 0 {
			url = options.Brokers[0]
		}
		return NewRabbitMQConsumer(
			WithRabbitMQConsumerURL(url),
			WithRabbitMQConsumerExchange(options.Topic),
			WithRabbitMQConsumerQueue(options.GroupID),
		)
	default:
		return nil, ErrMQNotSupported
	}
}

// ===== 示例代码 =====

// ExampleUsage 示例用法
// package main
//
// import (
//     "context"
//     "github.com/jiujuan/go-ants/pkg/mq"
// )
//
// func main() {
//     // 初始化 MQ 管理器
//     mq.InitMQ()
//
//     // 创建 Kafka 生产者
//     producer, err := mq.NewProducer(mq.MQTypeKafka,
//         mq.WithKafkaBrokers("localhost:9092"),
//         mq.WithKafkaTopic("my-topic"))
//     if err != nil {
//         panic(err)
//     }
//     defer producer.Close()
//
//     // 发送消息
//     ctx := context.Background()
//     msg := mq.NewBaseMessage("key", []byte("hello world"))
//     if err := producer.Send(ctx, msg); err != nil {
//         panic(err)
//     }
//
//     // 创建 RabbitMQ 消费者
//     consumer, err := mq.NewConsumer(mq.MQTypeRabbitMQ,
//         mq.WithRabbitMQConsumerURL("amqp://guest:guest@localhost:5672/"),
//         mq.WithRabbitMQConsumerExchange("my-exchange"),
//         mq.WithRabbitMQConsumerQueue("my-queue"))
//     if err != nil {
//         panic(err)
//     }
//     defer consumer.Close()
//
//     // 订阅消息
//     consumer.Subscribe(ctx, func(msg mq.Message) error {
//         println("received:", string(msg.GetValue()))
//         return nil
//     })
// }
