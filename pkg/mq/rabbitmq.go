package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== RabbitMQ 生产者 =====

// RabbitMQProducer RabbitMQ 生产者实现
type RabbitMQProducer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	opts    *RabbitMQProducerOptions
}

// RabbitMQProducerOption RabbitMQ 生产者选项函数
type RabbitMQProducerOption func(*RabbitMQProducerOptions)

type RabbitMQProducerOptions struct {
	URL          string
	Exchange     string
	ExchangeType string
	RoutingKey   string
	ContentType  string
	DeliveryMode amqp091.PublishingDeliveryMode
	Priority     uint8
	Mandatory    bool
	Immediate    bool
}

// WithRabbitMQURL 设置连接 URL
func WithRabbitMQURL(url string) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.URL = url
	}
}

// WithRabbitMQExchange 设置交换机
func WithRabbitMQExchange(exchange string) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.Exchange = exchange
	}
}

// WithRabbitMQExchangeType 设置交换机类型
func WithRabbitMQExchangeType(exchangeType string) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.ExchangeType = exchangeType
	}
}

// WithRabbitMQRoutingKey 设置路由键
func WithRabbitMQRoutingKey(routingKey string) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.RoutingKey = routingKey
	}
}

// WithRabbitMQContentType 设置内容类型
func WithRabbitMQContentType(contentType string) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.ContentType = contentType
	}
}

// WithRabbitMQDeliveryMode 设置传递模式
func WithRabbitMQDeliveryMode(mode amqp091.PublishingDeliveryMode) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.DeliveryMode = mode
	}
}

// WithRabbitMQPriority 设置优先级
func WithRabbitMQPriority(priority uint8) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.Priority = priority
	}
}

// WithRabbitMQMandatory 设置 mandatory
func WithRabbitMQMandatory(mandatory bool) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.Mandatory = mandatory
	}
}

// WithRabbitMQImmediate 设置 immediate
func WithRabbitMQImmediate(immediate bool) RabbitMQProducerOption {
	return func(o *RabbitMQProducerOptions) {
		o.Immediate = immediate
	}
}

// NewRabbitMQProducer 创建 RabbitMQ 生产者
func NewRabbitMQProducer(opts ...RabbitMQProducerOption) (*RabbitMQProducer, error) {
	options := &RabbitMQProducerOptions{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "default",
		ExchangeType: "direct",
		RoutingKey:   "default",
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Priority:     0,
		Mandatory:    false,
		Immediate:    false,
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

// Send 发送消息
func (p *RabbitMQProducer) Send(ctx context.Context, msg Message) error {
	publishing := amqp091.Publishing{
		ContentType:  p.opts.ContentType,
		DeliveryMode: p.opts.DeliveryMode,
		Priority:     p.opts.Priority,
		Body:         msg.GetValue(),
		Timestamp:    time.Now(),
	}

	// 设置 headers
	if len(msg.GetHeaders()) > 0 {
		publishing.Headers = make(amqp091.Table)
		for k, v := range msg.GetHeaders() {
			publishing.Headers[k] = v
		}
	}

	// 使用消息的 key 作为 routing key
	routingKey := msg.GetKey()
	if routingKey == "" {
		routingKey = p.opts.RoutingKey
	}

	return p.channel.PublishWithContext(
		ctx,
		p.opts.Exchange,
		routingKey,
		p.opts.Mandatory,
		p.opts.Immediate,
		publishing,
	)
}

// SendJSON 发送 JSON 消息
func (p *RabbitMQProducer) SendJSON(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := JSONMarshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	publishing := amqp091.Publishing{
		ContentType:  p.opts.ContentType,
		DeliveryMode: p.opts.DeliveryMode,
		Priority:     p.opts.Priority,
		Body:         data,
		Timestamp:    time.Now(),
	}

	routingKey := key
	if routingKey == "" {
		routingKey = p.opts.RoutingKey
	}

	return p.channel.PublishWithContext(
		ctx,
		p.opts.Exchange,
		routingKey,
		p.opts.Mandatory,
		p.opts.Immediate,
		publishing,
	)
}

// SendBatch 批量发送消息
func (p *RabbitMQProducer) SendBatch(ctx context.Context, msgs []Message) error {
	for _, msg := range msgs {
		if err := p.Send(ctx, msg); err != nil {
			return fmt.Errorf("failed to send batch message: %w", err)
		}
	}
	return nil
}

// Close 关闭生产者
func (p *RabbitMQProducer) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}

// ===== RabbitMQ 消费者 =====

// RabbitMQConsumer RabbitMQ 消费者实现
type RabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	opts    *RabbitMQConsumerOptions
}

// RabbitMQConsumerOption RabbitMQ 消费者选项函数
type RabbitMQConsumerOption func(*RabbitMQConsumerOptions)

type RabbitMQConsumerOptions struct {
	URL           string
	Exchange      string
	ExchangeType  string
	Queue         string
	RoutingKey    string
	ConsumerTag   string
	AutoAck       bool
	PrefetchCount int
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
}

// WithRabbitMQConsumerURL 设置连接 URL
func WithRabbitMQConsumerURL(url string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.URL = url
	}
}

// WithRabbitMQConsumerExchange 设置交换机
func WithRabbitMQConsumerExchange(exchange string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.Exchange = exchange
	}
}

// WithRabbitMQConsumerExchangeType 设置交换机类型
func WithRabbitMQConsumerExchangeType(exchangeType string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.ExchangeType = exchangeType
	}
}

// WithRabbitMQConsumerQueue 设置队列
func WithRabbitMQConsumerQueue(queue string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.Queue = queue
	}
}

// WithRabbitMQConsumerRoutingKey 设置路由键
func WithRabbitMQConsumerRoutingKey(routingKey string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.RoutingKey = routingKey
	}
}

// WithRabbitMQConsumerTag 设置消费者标签
func WithRabbitMQConsumerTag(tag string) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.ConsumerTag = tag
	}
}

// WithRabbitMQConsumerAutoAck 设置自动确认
func WithRabbitMQConsumerAutoAck(autoAck bool) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.AutoAck = autoAck
	}
}

// WithRabbitMQConsumerPrefetchCount 设置预取数量
func WithRabbitMQConsumerPrefetchCount(count int) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.PrefetchCount = count
	}
}

// WithRabbitMQConsumerExclusive 设置独占模式
func WithRabbitMQConsumerExclusive(exclusive bool) RabbitMQConsumerOption {
	return func(o *RabbitMQConsumerOptions) {
		o.Exclusive = exclusive
	}
}

// NewRabbitMQConsumer 创建 RabbitMQ 消费者
func NewRabbitMQConsumer(opts ...RabbitMQConsumerOption) (*RabbitMQConsumer, error) {
	options := &RabbitMQConsumerOptions{
		URL:           "amqp://guest:guest@localhost:5672/",
		Exchange:      "default",
		ExchangeType:  "direct",
		Queue:         "default",
		RoutingKey:    "default",
		ConsumerTag:   "",
		AutoAck:       false,
		PrefetchCount: 10,
		Exclusive:     false,
		NoLocal:       false,
		NoWait:        false,
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
		options.Exclusive,
		options.NoWait,
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

// Subscribe 订阅消息
func (c *RabbitMQConsumer) Subscribe(ctx context.Context, handler MessageHandler) error {
	deliveries, err := c.channel.Consume(
		c.opts.Queue,
		c.opts.ConsumerTag,
		c.opts.AutoAck,
		c.opts.Exclusive,
		c.opts.NoLocal,
		c.opts.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			msg := &RabbitMQConsumedMessage{
				routingKey: d.RoutingKey,
				body:       d.Body,
				headers:    make(map[string]interface{}),
				topic:      d.Exchange,
				timestamp:  d.Timestamp.Unix(),
			}

			// 转换 headers
			for k, v := range d.Headers {
				msg.headers[k] = v
			}

			if err := handler(msg); err != nil {
				log.Error("failed to handle message", log.Error(err))
				d.Nack(false, true) // requeue
			} else {
				d.Ack(false)
			}
		}
	}
}

// SubscribeTopic 订阅特定主题（RabbitMQ 使用 routing key 实现）
func (c *RabbitMQConsumer) SubscribeTopic(ctx context.Context, topic string, handler MessageHandler) error {
	// RabbitMQ 通过不同的 queue 来订阅不同主题
	// 这里创建一个新的 queue 绑定到指定的主题
	queueName := fmt.Sprintf("%s_%s", c.opts.Queue, topic)

	// 声明新队列
	queue, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		c.opts.Exclusive,
		c.opts.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定队列到指定的主题
	err = c.channel.QueueBind(
		queue.Name,
		topic, // 使用 topic 作为 routing key
		c.opts.Exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	deliveries, err := c.channel.Consume(
		queue.Name,
		c.opts.ConsumerTag,
		c.opts.AutoAck,
		c.opts.Exclusive,
		c.opts.NoLocal,
		c.opts.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			msg := &RabbitMQConsumedMessage{
				routingKey: d.RoutingKey,
				body:       d.Body,
				headers:    make(map[string]interface{}),
				topic:      d.Exchange,
				timestamp:  d.Timestamp.Unix(),
			}

			for k, v := range d.Headers {
				msg.headers[k] = v
			}

			if err := handler(msg); err != nil {
				log.Error("failed to handle message", log.Error(err))
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}
}

// Close 关闭消费者
func (c *RabbitMQConsumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

// ===== RabbitMQ 消费消息实现 =====

// RabbitMQConsumedMessage RabbitMQ 消费的消息实现
type RabbitMQConsumedMessage struct {
	routingKey string
	body       []byte
	headers    map[string]interface{}
	topic      string
	timestamp  int64
}

func (m *RabbitMQConsumedMessage) GetKey() string                     { return m.routingKey }
func (m *RabbitMQConsumedMessage) GetValue() []byte                   { return m.body }
func (m *RabbitMQConsumedMessage) GetHeaders() map[string]interface{} { return m.headers }
func (m *RabbitMQConsumedMessage) GetTopic() string                   { return m.topic }
func (m *RabbitMQConsumedMessage) GetTimestamp() int64                { return m.timestamp }

// ===== RabbitMQ 工厂 =====

// RabbitMQFactory RabbitMQ 工厂实现
type RabbitMQFactory struct{}

// NewRabbitMQFactory 创建 RabbitMQ 工厂
func NewRabbitMQFactory() *RabbitMQFactory {
	return &RabbitMQFactory{}
}

// CreateProducer 创建生产者
func (f *RabbitMQFactory) CreateProducer(opts ...ProducerOption) (Producer, error) {
	producerOpts := &ProducerOptions{}
	for _, opt := range opts {
		opt(producerOpts)
	}

	return NewRabbitMQProducer()
}

// CreateConsumer 创建消费者
func (f *RabbitMQFactory) CreateConsumer(opts ...ConsumerOption) (Consumer, error) {
	consumerOpts := &ConsumerOptions{}
	for _, opt := range opts {
		opt(consumerOpts)
	}

	return NewRabbitMQConsumer()
}

// GetMQType 获取 MQ 类型
func (f *RabbitMQFactory) GetMQType() MQType {
	return MQTypeRabbitMQ
}
