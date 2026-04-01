package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== Kafka 生产者 =====

// KafkaProducer Kafka 生产者实现
type KafkaProducer struct {
	writer *kafka.Writer
	opts   *KafkaProducerOptions
}

// KafkaProducerOption Kafka 生产者选项函数
type KafkaProducerOption func(*KafkaProducerOptions)

type KafkaProducerOptions struct {
	Brokers       []string
	Topic         string
	Balancer      *kafka.Balancer
	BatchSize     int
	BatchTimeout  time.Duration
	Async         bool
	Compression   kafka.Compression
	MaxAttempts   int
	RetryInterval time.Duration
}

// WithKafkaBrokers 设置 broker 列表
func WithKafkaBrokers(brokers ...string) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.Brokers = brokers
	}
}

// WithKafkaTopic 设置主题
func WithKafkaTopic(topic string) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.Topic = topic
	}
}

// WithKafkaBalancer 设置负载均衡器
func WithKafkaBalancer(balancer *kafka.Balancer) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.Balancer = balancer
	}
}

// WithKafkaBatchSize 设置批量大小
func WithKafkaBatchSize(batchSize int) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.BatchSize = batchSize
	}
}

// WithKafkaBatchTimeout 设置批量超时
func WithKafkaBatchTimeout(timeout time.Duration) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.BatchTimeout = timeout
	}
}

// WithKafkaAsync 设置异步模式
func WithKafkaAsync(async bool) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.Async = async
	}
}

// WithKafkaCompression 设置压缩方式
func WithKafkaCompression(compression kafka.Compression) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.Compression = compression
	}
}

// WithKafkaMaxAttempts 设置最大重试次数
func WithKafkaMaxAttempts(maxAttempts int) KafkaProducerOption {
	return func(o *KafkaProducerOptions) {
		o.MaxAttempts = maxAttempts
	}
}

// NewKafkaProducer 创建 Kafka 生产者
func NewKafkaProducer(opts ...KafkaProducerOption) (*KafkaProducer, error) {
	options := &KafkaProducerOptions{
		Brokers:      []string{"localhost:9092"},
		Topic:        "default",
		BatchSize:    10,
		BatchTimeout: 10 * time.Millisecond,
		Compression:  kafka.Snappy,
		MaxAttempts:  3,
	}

	for _, opt := range opts {
		opt(options)
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(options.Brokers...),
		Topic:        options.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    options.BatchSize,
		BatchTimeout: options.BatchTimeout,
		Async:        options.Async,
		Compression:  options.Compression,
		MaxAttempts:  options.MaxAttempts,
	}

	log.Info("kafka producer created",
		log.String("brokers", fmt.Sprintf("%v", options.Brokers)),
		log.String("topic", options.Topic))

	return &KafkaProducer{
		writer: writer,
		opts:   options,
	}, nil
}

// Send 发送消息
func (p *KafkaProducer) Send(ctx context.Context, msg Message) error {
	kafkaMsg := kafka.Message{
		Key:   []byte(msg.GetKey()),
		Value: msg.GetValue(),
	}

	// 设置 headers
	if len(msg.GetHeaders()) > 0 {
		kafkaMsg.Headers = make([]kafka.Header, 0, len(msg.GetHeaders()))
		for k, v := range msg.GetHeaders() {
			// 简单转换，实际使用时可根据需要调整
			kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
				Key:   k,
				Value: []byte(fmt.Sprintf("%v", v)),
			})
		}
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// SendJSON 发送 JSON 消息
func (p *KafkaProducer) SendJSON(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := JSONMarshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}

	if topic != "" {
		msg.Topic = topic
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to produce json message: %w", err)
	}

	return nil
}

// SendBatch 批量发送消息
func (p *KafkaProducer) SendBatch(ctx context.Context, msgs []Message) error {
	if len(msgs) == 0 {
		return nil
	}

	kafkaMsgs := make([]kafka.Message, 0, len(msgs))
	for _, msg := range msgs {
		kafkaMsg := kafka.Message{
			Key:   []byte(msg.GetKey()),
			Value: msg.GetValue(),
		}
		kafkaMsgs = append(kafkaMsgs, kafkaMsg)
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
		return fmt.Errorf("failed to produce batch messages: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// ===== Kafka 消费者 =====

// KafkaConsumer Kafka 消费者实现
type KafkaConsumer struct {
	reader *kafka.Reader
	opts   *KafkaConsumerOptions
}

// KafkaConsumerOption Kafka 消费者选项函数
type KafkaConsumerOption func(*KafkaConsumerOptions)

type KafkaConsumerOptions struct {
	Brokers     []string
	Topic       string
	GroupID     string
	MinBytes    int
	MaxBytes    int
	MaxAttempts int
	StartOffset int64
}

// WithKafkaConsumerBrokers 设置 broker 列表
func WithKafkaConsumerBrokers(brokers ...string) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.Brokers = brokers
	}
}

// WithKafkaConsumerTopic 设置主题
func WithKafkaConsumerTopic(topic string) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.Topic = topic
	}
}

// WithKafkaConsumerGroupID 设置消费者组 ID
func WithKafkaConsumerGroupID(groupID string) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.GroupID = groupID
	}
}

// WithKafkaConsumerMinBytes 设置最小字节数
func WithKafkaConsumerMinBytes(minBytes int) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.MinBytes = minBytes
	}
}

// WithKafkaConsumerMaxBytes 设置最大字节数
func WithKafkaConsumerMaxBytes(maxBytes int) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.MaxBytes = maxBytes
	}
}

// WithKafkaConsumerMaxAttempts 设置最大重试次数
func WithKafkaConsumerMaxAttempts(maxAttempts int) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.MaxAttempts = maxAttempts
	}
}

// WithKafkaConsumerStartOffset 设置起始偏移量
func WithKafkaConsumerStartOffset(startOffset int64) KafkaConsumerOption {
	return func(o *KafkaConsumerOptions) {
		o.StartOffset = startOffset
	}
}

// NewKafkaConsumer 创建 Kafka 消费者
func NewKafkaConsumer(opts ...KafkaConsumerOption) (*KafkaConsumer, error) {
	options := &KafkaConsumerOptions{
		Brokers:     []string{"localhost:9092"},
		Topic:       "default",
		GroupID:     "default-group",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxAttempts: 3,
		StartOffset: kafka.LastOffset,
	}

	for _, opt := range opts {
		opt(options)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     options.Brokers,
		Topic:       options.Topic,
		GroupID:     options.GroupID,
		MinBytes:    options.MinBytes,
		MaxBytes:    options.MaxBytes,
		MaxAttempts: options.MaxAttempts,
		StartOffset: options.StartOffset,
	})

	log.Info("kafka consumer created",
		log.String("brokers", fmt.Sprintf("%v", options.Brokers)),
		log.String("topic", options.Topic),
		log.String("groupID", options.GroupID))

	return &KafkaConsumer{
		reader: reader,
		opts:   options,
	}, nil
}

// Subscribe 订阅消息
func (c *KafkaConsumer) Subscribe(ctx context.Context, handler MessageHandler) error {
	return c.subscribeLoop(ctx, handler)
}

// SubscribeTopic 订阅特定主题
func (c *KafkaConsumer) SubscribeTopic(ctx context.Context, topic string, handler MessageHandler) error {
	// Kafka 按主题消费需要创建新的 reader
	if topic != c.opts.Topic {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     c.opts.Brokers,
			Topic:       topic,
			GroupID:     c.opts.GroupID,
			MinBytes:    c.opts.MinBytes,
			MaxBytes:    c.opts.MaxBytes,
			MaxAttempts: c.opts.MaxAttempts,
			StartOffset: c.opts.StartOffset,
		})
		defer reader.Close()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					log.Error("failed to read message", log.Error(err))
					continue
				}

				kafkaMsg := &KafkaConsumedMessage{
					key:       string(msg.Key),
					value:     msg.Value,
					headers:   make(map[string]interface{}),
					topic:     msg.Topic,
					timestamp: msg.Time.Unix(),
				}

				if err := handler(kafkaMsg); err != nil {
					log.Error("failed to handle message", log.Error(err))
				}

				if err := reader.CommitMessages(ctx, msg); err != nil {
					log.Error("failed to commit message", log.Error(err))
				}
			}
		}
	}

	return c.subscribeLoop(ctx, handler)
}

// subscribeLoop 内部订阅循环
func (c *KafkaConsumer) subscribeLoop(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Error("failed to read message", log.Error(err))
				continue
			}

			kafkaMsg := &KafkaConsumedMessage{
				key:       string(msg.Key),
				value:     msg.Value,
				headers:   make(map[string]interface{}),
				topic:     msg.Topic,
				timestamp: msg.Time.Unix(),
			}

			if err := handler(kafkaMsg); err != nil {
				log.Error("failed to handle message", log.Error(err))
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Error("failed to commit message", log.Error(err))
			}
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// ===== Kafka 消费消息实现 =====

// KafkaConsumedMessage Kafka 消费的消息实现
type KafkaConsumedMessage struct {
	key       string
	value     []byte
	headers   map[string]interface{}
	topic     string
	timestamp int64
}

func (m *KafkaConsumedMessage) GetKey() string                     { return m.key }
func (m *KafkaConsumedMessage) GetValue() []byte                   { return m.value }
func (m *KafkaConsumedMessage) GetHeaders() map[string]interface{} { return m.headers }
func (m *KafkaConsumedMessage) GetTopic() string                   { return m.topic }
func (m *KafkaConsumedMessage) GetTimestamp() int64                { return m.timestamp }

// ===== Kafka 工厂 =====

// KafkaFactory Kafka 工厂实现
type KafkaFactory struct{}

// NewKafkaFactory 创建 Kafka 工厂
func NewKafkaFactory() *KafkaFactory {
	return &KafkaFactory{}
}

// CreateProducer 创建生产者
func (f *KafkaFactory) CreateProducer(opts ...ProducerOption) (Producer, error) {
	producerOpts := &ProducerOptions{}
	for _, opt := range opts {
		opt(producerOpts)
	}

	// 使用默认选项创建生产者
	return NewKafkaProducer()
}

// CreateConsumer 创建消费者
func (f *KafkaFactory) CreateConsumer(opts ...ConsumerOption) (Consumer, error) {
	consumerOpts := &ConsumerOptions{}
	for _, opt := range opts {
		opt(consumerOpts)
	}

	// 使用默认选项创建消费者
	return NewKafkaConsumer()
}

// GetMQType 获取 MQ 类型
func (f *KafkaFactory) GetMQType() MQType {
	return MQTypeKafka
}
