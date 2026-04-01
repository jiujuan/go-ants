package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"

	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== Kafka =====

// KafkaProducer Kafka 生产者
type KafkaProducer struct {
	writer *kafka.Writer
	opts   *KafkaOptions
}

// KafkaConsumer Kafka 消费者
type KafkaConsumer struct {
	reader *kafka.Reader
	opts   *KafkaOptions
}

// KafkaOption Kafka 选项函数
type KafkaOption func(*KafkaOptions)

type KafkaOptions struct {
	Brokers       []string
	Topic         string
	GroupID       string
	Balancer      *kafka.Balancer
	BatchSize     int
	BatchTimeout  time.Duration
	Async         bool
	Compression   kafka.Compression
	MaxAttempts   int
	RetryInterval time.Duration
}

// WithKafkaBrokers 设置 broker 列表
func WithKafkaBrokers(brokers ...string) KafkaOption {
	return func(o *KafkaOptions) {
		o.Brokers = brokers
	}
}

// WithKafkaTopic 设置主题
func WithKafkaTopic(topic string) KafkaOption {
	return func(o *KafkaOptions) {
		o.Topic = topic
	}
}

// WithKafkaGroupID 设置消费者组 ID
func WithKafkaGroupID(groupID string) KafkaOption {
	return func(o *KafkaOptions) {
		o.GroupID = groupID
	}
}

// NewKafkaProducer 创建 Kafka 生产者
func NewKafkaProducer(opts ...KafkaOption) (*KafkaProducer, error) {
	options := &KafkaOptions{
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

// Produce 发送消息
func (p *KafkaProducer) Produce(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// ProduceJSON 发送 JSON 消息
func (p *KafkaProducer) ProduceJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	return p.Produce(ctx, []byte(key), data)
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// NewKafkaConsumer 创建 Kafka 消费者
func NewKafkaConsumer(opts ...KafkaOption) (*KafkaConsumer, error) {
	options := &KafkaOptions{
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

// Consume 消费消息
func (c *KafkaConsumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Error("failed to read message",
					log.Error(err))
				continue
			}

			if err := handler(msg.Key, msg.Value); err != nil {
				log.Error("failed to handle message",
					log.Error(err))
			}
		}
	}
}

// ConsumeLoop 消费消息循环
func (c *KafkaConsumer) ConsumeLoop(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Error("failed to fetch message",
				log.Error(err))
			continue
		}

		if err := handler(msg.Key, msg.Value); err != nil {
			log.Error("failed to handle message",
				log.Error(err))
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Error("failed to commit message",
				log.Error(err))
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// ===== RabbitMQ =====

// RabbitMQProducer RabbitMQ 生产者
type RabbitMQProducer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	opts    *RabbitMQOptions
}

// RabbitMQConsumer RabbitMQ 消费者
type RabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	opts    *RabbitMQOptions
}

// RabbitMQOption RabbitMQ 选项函数
type RabbitMQOption func(*RabbitMQOptions)

type RabbitMQOptions struct {
	URL           string
	Exchange      string
	ExchangeType  string
	Queue         string
	RoutingKey    string
	ConsumerTag   string
	AutoAck       bool
	PrefetchCount int
	ContentType   string
	DeliveryMode  amqp091.PublishingDeliveryMode
	Priority      uint8
}

// WithRabbitMQURL 设置连接 URL
func WithRabbitMQURL(url string) RabbitMQOption {
	return func(o *RabbitMQOptions) {
		o.URL = url
	}
}

// WithRabbitMQExchange 设置交换机
func WithRabbitMQExchange(exchange string) RabbitMQOption {
	return func(o *RabbitMQOptions) {
		o.Exchange = exchange
	}
}

// WithRabbitMQExchangeType 设置交换机类型
func WithRabbitMQExchangeType(exchangeType string) RabbitMQOption {
	return func(o *RabbitMQOptions) {
		o.ExchangeType = exchangeType
	}
}

// WithRabbitMQQueue 设置队列
func WithRabbitMQQueue(queue string) RabbitMQOption {
	return func(o *RabbitMQOptions) {
		o.Queue = queue
	}
}

// WithRabbitMQRoutingKey 设置路由键
func WithRabbitMQRoutingKey(routingKey string) RabbitMQOption {
	return func(o *RabbitMQOptions) {
		o.RoutingKey = routingKey
	}
}

// NewRabbitMQProducer 创建 RabbitMQ 生产者
func NewRabbitMQProducer(opts ...RabbitMQOption) (*RabbitMQProducer, error) {
	options := &RabbitMQOptions{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "default",
		ExchangeType: "direct",
		RoutingKey:   "default",
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
	}

	for _, opt := range opts {
		opt(options)
	}

	conn, err := amqp091.Dial(options.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明交换机
	err = ch.ExchangeDeclare(
		options.Exchange,
		options.ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Info("rabbitmq producer created",
		log.String("url", options.URL),
		log.String("exchange", options.Exchange))

	return &RabbitMQProducer{
		conn:    conn,
		channel: ch,
		opts:    options,
	}, nil
}

// Publish 发布消息
func (p *RabbitMQProducer) Publish(ctx context.Context, routingKey string, body []byte) error {
	msg := amqp091.Publishing{
		ContentType:  p.opts.ContentType,
		DeliveryMode: p.opts.DeliveryMode,
		Priority:     p.opts.Priority,
		Body:         body,
		Timestamp:    time.Now(),
	}

	if routingKey == "" {
		routingKey = p.opts.RoutingKey
	}

	return p.channel.PublishWithContext(
		ctx,
		p.opts.Exchange,
		routingKey,
		false,
		false,
		msg,
	)
}

// PublishJSON 发布 JSON 消息
func (p *RabbitMQProducer) PublishJSON(ctx context.Context, routingKey string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	return p.Publish(ctx, routingKey, data)
}

// Close 关闭生产者
func (p *RabbitMQProducer) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}

// NewRabbitMQConsumer 创建 RabbitMQ 消费者
func NewRabbitMQConsumer(opts ...RabbitMQOption) (*RabbitMQConsumer, error) {
	options := &RabbitMQOptions{
		URL:           "amqp://guest:guest@localhost:5672/",
		Exchange:      "default",
		ExchangeType:  "direct",
		Queue:         "default",
		RoutingKey:    "default",
		AutoAck:       false,
		PrefetchCount: 10,
	}

	for _, opt := range opts {
		opt(options)
	}

	conn, err := amqp091.Dial(options.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 设置 QoS
	err = ch.Qos(options.PrefetchCount, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// 声明交换机
	err = ch.ExchangeDeclare(
		options.Exchange,
		options.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 声明队列
	queue, err := ch.QueueDeclare(
		options.Queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定队列
	err = ch.QueueBind(
		queue.Name,
		options.RoutingKey,
		options.Exchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Info("rabbitmq consumer created",
		log.String("url", options.URL),
		log.String("exchange", options.Exchange),
		log.String("queue", options.Queue))

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		opts:    options,
	}, nil
}

// Consume 消费消息
func (c *RabbitMQConsumer) Consume(handler func(body []byte) error) error {
	deliveries, err := c.channel.Consume(
		c.opts.Queue,
		c.opts.ConsumerTag,
		c.opts.AutoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for d := range deliveries {
		if err := handler(d.Body); err != nil {
			log.Error("failed to handle message",
				log.Error(err))
			d.Nack(false, true) // requeue
		} else {
			d.Ack(false)
		}
	}

	return nil
}

// Close 关闭消费者
func (c *RabbitMQConsumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

// ===== 通用消息接口 =====

// Message 消息接口
type Message interface {
	GetKey() string
	GetValue() []byte
	GetHeaders() map[string]interface{}
}

// KafkaMessage Kafka 消息实现
type KafkaMessage struct {
	key     string
	value   []byte
	headers map[string]interface{}
}

func NewKafkaMessage(key string, value []byte) *KafkaMessage {
	return &KafkaMessage{
		key:     key,
		value:   value,
		headers: make(map[string]interface{}),
	}
}

func (m *KafkaMessage) GetKey() string                     { return m.key }
func (m *KafkaMessage) GetValue() []byte                   { return m.value }
func (m *KafkaMessage) GetHeaders() map[string]interface{} { return m.headers }

// WithMessageHeader 设置消息头
func WithMessageHeader(key string, value interface{}) func(*KafkaMessage) {
	return func(m *KafkaMessage) {
		m.headers[key] = value
	}
}
