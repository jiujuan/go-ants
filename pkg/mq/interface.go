package mq

import (
	"context"
	"encoding/json"
	"fmt"
)

// ===== 消息队列类型 =====

// MQType 消息队列类型枚举
type MQType string

const (
	MQTypeKafka    MQType = "kafka"
	MQTypeRabbitMQ MQType = "rabbitmq"
	MQTypeRocketMQ MQType = "rocketmq"
)

// ===== 通用消息接口 =====

// Message 消息接口，所有消息实现必须实现此接口
type Message interface {
	// GetKey 获取消息键
	GetKey() string
	// GetValue 获取消息值
	GetValue() []byte
	// GetHeaders 获取消息头
	GetHeaders() map[string]interface{}
	// GetTopic 获取消息主题（部分 MQ 支持）
	GetTopic() string
	// GetTimestamp 获取消息时间戳
	GetTimestamp() int64
}

// BaseMessage 基础消息实现
type BaseMessage struct {
	key       string
	value     []byte
	headers   map[string]interface{}
	topic     string
	timestamp int64
}

// NewBaseMessage 创建基础消息
func NewBaseMessage(key string, value []byte) *BaseMessage {
	return &BaseMessage{
		key:       key,
		value:     value,
		headers:   make(map[string]interface{}),
		timestamp: 0,
	}
}

func (m *BaseMessage) GetKey() string                     { return m.key }
func (m *BaseMessage) GetValue() []byte                   { return m.value }
func (m *BaseMessage) GetHeaders() map[string]interface{} { return m.headers }
func (m *BaseMessage) GetTopic() string                   { return m.topic }
func (m *BaseMessage) GetTimestamp() int64                { return m.timestamp }

// WithMessageTopic 设置消息主题
func WithMessageTopic(topic string) func(*BaseMessage) {
	return func(m *BaseMessage) {
		m.topic = topic
	}
}

// WithMessageTimestamp 设置消息时间戳
func WithMessageTimestamp(ts int64) func(*BaseMessage) {
	return func(m *BaseMessage) {
		m.timestamp = ts
	}
}

// WithMessageHeader 设置消息头
func WithMessageHeader(key string, value interface{}) func(*BaseMessage) {
	return func(m *BaseMessage) {
		m.headers[key] = value
	}
}

// ===== 生产者接口 =====

// Producer 消息生产者接口，所有 MQ 生产者实现必须实现此接口
type Producer interface {
	// Send 发送消息
	Send(ctx context.Context, msg Message) error
	// SendJSON 发送 JSON 消息
	SendJSON(ctx context.Context, topic string, key string, value interface{}) error
	// SendBatch 批量发送消息
	SendBatch(ctx context.Context, msgs []Message) error
	// Close 关闭生产者
	Close() error
}

// ProducerOptions 生产者选项
type ProducerOptions struct {
	MQType MQType
}

// ProducerOption 生产者选项函数类型
type ProducerOption func(*ProducerOptions)

// WithMQType 设置 MQ 类型
func WithMQType(mqType MQType) ProducerOption {
	return func(o *ProducerOptions) {
		o.MQType = mqType
	}
}

// ===== 消费者接口 =====

// MessageHandler 消息处理函数类型
type MessageHandler func(msg Message) error

// Consumer 消息消费者接口，所有 MQ 消费者实现必须实现此接口
type Consumer interface {
	// Subscribe 订阅消息
	Subscribe(ctx context.Context, handler MessageHandler) error
	// SubscribeTopic 订阅特定主题（部分 MQ 支持）
	SubscribeTopic(ctx context.Context, topic string, handler MessageHandler) error
	// Close 关闭消费者
	Close() error
}

// ConsumerOptions 消费者选项
type ConsumerOptions struct {
	MQType     MQType
	GroupID    string
	AutoCommit bool
}

// ConsumerOption 消费者选项函数类型
type ConsumerOption func(*ConsumerOptions)

// WithConsumerGroupID 设置消费者组 ID
func WithConsumerGroupID(groupID string) ConsumerOption {
	return func(o *ConsumerOptions) {
		o.GroupID = groupID
	}
}

// WithConsumerAutoCommit 设置自动提交
func WithConsumerAutoCommit(autoCommit bool) ConsumerOption {
	return func(o *ConsumerOptions) {
		o.AutoCommit = autoCommit
	}
}

// ===== MQ 工厂接口 =====

// MQFactory 消息队列工厂接口
type MQFactory interface {
	// CreateProducer 创建生产者
	CreateProducer(opts ...ProducerOption) (Producer, error)
	// CreateConsumer 创建消费者
	CreateConsumer(opts ...ConsumerOption) (Consumer, error)
	// GetMQType 获取 MQ 类型
	GetMQType() MQType
}

// ===== 公共工具函数 =====

// JSONMarshal JSON 序列化
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JSONUnmarshal JSON 反序列化
func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewMessageFromBytes 从字节数组创建消息
func NewMessageFromBytes(key string, value []byte, opts ...func(*BaseMessage)) Message {
	msg := NewBaseMessage(key, value)
	for _, opt := range opts {
		opt(msg)
	}
	return msg
}

// NewJSONMessage 创建 JSON 消息
func NewJSONMessage(key string, value interface{}, opts ...func(*BaseMessage)) (Message, error) {
	data, err := JSONMarshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}
	msg := NewBaseMessage(key, data)
	for _, opt := range opts {
		opt(msg)
	}
	return msg, nil
}

// ===== 错误定义 =====

var (
	ErrMQNotSupported     = fmt.Errorf("mq type not supported")
	ErrConnectionFailed   = fmt.Errorf("mq connection failed")
	ErrProduceFailed      = fmt.Errorf("mq produce message failed")
	ErrConsumeFailed      = fmt.Errorf("mq consume message failed")
	ErrCloseFailed        = fmt.Errorf("mq close failed")
	ErrInvalidMessage     = fmt.Errorf("invalid message")
	ErrSubscriptionFailed = fmt.Errorf("mq subscription failed")
)
