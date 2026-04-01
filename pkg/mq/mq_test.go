package mq

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// ===== Interface Tests =====

func TestMQType_Constants(t *testing.T) {
	tests := []struct {
		mqType   MQType
		expected string
	}{
		{MQTypeKafka, "kafka"},
		{MQTypeRabbitMQ, "rabbitmq"},
		{MQTypeRocketMQ, "rocketmq"},
	}

	for _, tt := range tests {
		if string(tt.mqType) != tt.expected {
			t.Errorf("MQType 期望 %s, 得到 %s", tt.expected, tt.mqType)
		}
	}
}

func TestNewBaseMessage(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"))

	if msg.GetKey() != "key" {
		t.Errorf("消息 Key 期望 'key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "value" {
		t.Errorf("消息 Value 期望 'value', 得到 %s", string(msg.GetValue()))
	}

	if msg.GetHeaders() == nil {
		t.Error("消息 Headers 不应该为 nil")
	}

	if msg.GetTopic() != "" {
		t.Errorf("消息 Topic 期望为空, 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != 0 {
		t.Errorf("消息 Timestamp 期望 0, 得到 %d", msg.GetTimestamp())
	}
}

func TestWithMessageTopic(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"), WithMessageTopic("test-topic"))

	if msg.GetTopic() != "test-topic" {
		t.Errorf("消息 Topic 期望 'test-topic', 得到 %s", msg.GetTopic())
	}
}

func TestWithMessageTimestamp(t *testing.T) {
	timestamp := int64(1234567890)
	msg := NewBaseMessage("key", []byte("value"), WithMessageTimestamp(timestamp))

	if msg.GetTimestamp() != timestamp {
		t.Errorf("消息 Timestamp 期望 %d, 得到 %d", timestamp, msg.GetTimestamp())
	}
}

func TestWithMessageHeader(t *testing.T) {
	msg := NewBaseMessage("key", []byte("value"), WithMessageHeader("header-key", "header-value"))

	headers := msg.GetHeaders()
	if headers["header-key"] != "header-value" {
		t.Errorf("消息 Header 期望 'header-value', 得到 %v", headers["header-key"])
	}
}

func TestBaseMessage_AllOptions(t *testing.T) {
	timestamp := int64(1234567890)
	msg := NewBaseMessage(
		"my-key",
		[]byte("my-value"),
		WithMessageTopic("my-topic"),
		WithMessageTimestamp(timestamp),
		WithMessageHeader("key1", "value1"),
		WithMessageHeader("key2", "value2"),
	)

	if msg.GetKey() != "my-key" {
		t.Errorf("Key 期望 'my-key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "my-value" {
		t.Errorf("Value 期望 'my-value', 得到 %s", string(msg.GetValue()))
	}

	if msg.GetTopic() != "my-topic" {
		t.Errorf("Topic 期望 'my-topic', 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != timestamp {
		t.Errorf("Timestamp 期望 %d, 得到 %d", timestamp, msg.GetTimestamp())
	}

	headers := msg.GetHeaders()
	if len(headers) != 2 {
		t.Errorf("Headers 长度期望 2, 得到 %d", len(headers))
	}

	if headers["key1"] != "value1" {
		t.Errorf("Header key1 期望 'value1', 得到 %v", headers["key1"])
	}
}

// ===== Producer Options Tests =====

func TestProducerOptions(t *testing.T) {
	opts := &ProducerOptions{}

	WithMQType(MQTypeKafka)(opts)

	if opts.MQType != MQTypeKafka {
		t.Errorf("MQType 期望 Kafka, 得到 %s", opts.MQType)
	}
}

func TestProducerOptions_Defaults(t *testing.T) {
	opts := &ProducerOptions{}

	if opts.MQType != "" {
		t.Errorf("默认 MQType 应该为空, 得到 %s", opts.MQType)
	}
}

// ===== Consumer Options Tests =====

func TestConsumerOptions(t *testing.T) {
	opts := &ConsumerOptions{}

	WithConsumerGroupID("test-group")(opts)
	WithConsumerAutoCommit(true)(opts)

	if opts.GroupID != "test-group" {
		t.Errorf("GroupID 期望 'test-group', 得到 %s", opts.GroupID)
	}

	if !opts.AutoCommit {
		t.Error("AutoCommit 应该为 true")
	}
}

func TestConsumerOptions_Defaults(t *testing.T) {
	opts := &ConsumerOptions{}

	if opts.GroupID != "" {
		t.Errorf("默认 GroupID 应该为空, 得到 %s", opts.GroupID)
	}

	if opts.AutoCommit {
		t.Error("默认 AutoCommit 应该为 false")
	}
}

// ===== Message Handler Tests =====

func TestMessageHandler_Type(t *testing.T) {
	handler := func(msg Message) error {
		return nil
	}

	if handler == nil {
		t.Error("MessageHandler 不应该为 nil")
	}
}

func TestMessageHandler_WithError(t *testing.T) {
	testErr := NewTestError("test error")
	handler := func(msg Message) error {
		return testErr
	}

	msg := NewBaseMessage("key", []byte("value"))
	err := handler(msg)

	if err != testErr {
		t.Errorf("Handler 应该返回测试错误")
	}
}

// ===== JSON Marshal/Unmarshal Tests =====

func TestJSONMarshal_Success(t *testing.T) {
	data := map[string]string{"key": "value"}

	result, err := JSONMarshal(data)
	if err != nil {
		t.Errorf("JSONMarshal 不应该返回错误: %v", err)
	}

	if len(result) == 0 {
		t.Error("JSONMarshal 结果不应该为空")
	}
}

func TestJSONMarshal_Struct(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	s := TestStruct{Name: "John", Age: 30}
	result, err := JSONMarshal(s)

	if err != nil {
		t.Errorf("JSONMarshal 不应该返回错误: %v", err)
	}

	var decoded TestStruct
	err = json.Unmarshal(result, &decoded)
	if err != nil {
		t.Errorf("JSONUnmarshal 不应该返回错误: %v", err)
	}

	if decoded.Name != "John" {
		t.Errorf("解码后 Name 期望 'John', 得到 %s", decoded.Name)
	}

	if decoded.Age != 30 {
		t.Errorf("解码后 Age 期望 30, 得到 %d", decoded.Age)
	}
}

func TestJSONUnmarshal_Success(t *testing.T) {
	data := []byte(`{"name":"Alice","age":25}`)

	var result map[string]interface{}
	err := JSONUnmarshal(data, &result)
	if err != nil {
		t.Errorf("JSONUnmarshal 不应该返回错误: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("解码后 name 期望 'Alice', 得到 %v", result["name"])
	}
}

func TestJSONUnmarshal_InvalidJSON(t *testing.T) {
	data := []byte(`invalid json`)

	var result map[string]interface{}
	err := JSONUnmarshal(data, &result)
	if err == nil {
		t.Error("JSONUnmarshal 应该返回错误（无效 JSON）")
	}
}

// ===== NewMessageFromBytes Tests =====

func TestNewMessageFromBytes(t *testing.T) {
	msg := NewMessageFromBytes("key", []byte("value"))

	if msg.GetKey() != "key" {
		t.Errorf("Key 期望 'key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "value" {
		t.Errorf("Value 期望 'value', 得到 %s", string(msg.GetValue()))
	}
}

func TestNewMessageFromBytes_WithOptions(t *testing.T) {
	msg := NewMessageFromBytes(
		"key",
		[]byte("value"),
		WithMessageTopic("topic"),
		WithMessageHeader("h", "v"),
	)

	if msg.GetTopic() != "topic" {
		t.Errorf("Topic 期望 'topic', 得到 %s", msg.GetTopic())
	}

	if msg.GetHeaders()["h"] != "v" {
		t.Error("Header 设置失败")
	}
}

// ===== NewJSONMessage Tests =====

func TestNewJSONMessage_Success(t *testing.T) {
	data := map[string]string{"key": "value"}

	msg, err := NewJSONMessage("key", data)
	if err != nil {
		t.Errorf("NewJSONMessage 不应该返回错误: %v", err)
	}

	if msg.GetKey() != "key" {
		t.Errorf("Key 期望 'key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) == "" {
		t.Error("Value 不应该为空")
	}
}

func TestNewJSONMessage_WithOptions(t *testing.T) {
	msg, err := NewJSONMessage(
		"key",
		"value",
		WithMessageTopic("topic"),
		WithMessageTimestamp(12345),
	)

	if err != nil {
		t.Errorf("NewJSONMessage 不应该返回错误: %v", err)
	}

	if msg.GetTopic() != "topic" {
		t.Errorf("Topic 期望 'topic', 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != 12345 {
		t.Errorf("Timestamp 期望 12345, 得到 %d", msg.GetTimestamp())
	}
}

func TestNewJSONMessage_WithStruct(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{Name: "Bob", Age: 35}
	msg, err := NewJSONMessage("user-key", user)

	if err != nil {
		t.Errorf("NewJSONMessage 不应该返回错误: %v", err)
	}

	// 验证 JSON 内容
	var decoded User
	err = json.Unmarshal(msg.GetValue(), &decoded)
	if err != nil {
		t.Errorf("JSON 解码失败: %v", err)
	}

	if decoded.Name != "Bob" {
		t.Errorf("解码后 Name 期望 'Bob', 得到 %s", decoded.Name)
	}
}

// ===== Error Tests =====

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrMQNotSupported, "mq type not supported"},
		{ErrConnectionFailed, "mq connection failed"},
		{ErrProduceFailed, "mq produce message failed"},
		{ErrConsumeFailed, "mq consume message failed"},
		{ErrCloseFailed, "mq close failed"},
		{ErrInvalidMessage, "invalid message"},
		{ErrSubscriptionFailed, "mq subscription failed"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expected {
			t.Errorf("错误消息期望 '%s', 得到 '%s'", tt.expected, tt.err.Error())
		}
	}
}

// ===== Helper Types for Testing =====

type TestError struct {
	msg string
}

func NewTestError(msg string) *TestError {
	return &TestError{msg: msg}
}

func (e *TestError) Error() string {
	return e.msg
}

// ===== Interface Compliance Tests =====

func TestMessageInterface_Compliance(t *testing.T) {
	// 确保 BaseMessage 实现了 Message 接口
	var _ Message = &BaseMessage{}

	msg := NewBaseMessage("test-key", []byte("test-value"))

	// 测试所有接口方法
	_ = msg.GetKey()
	_ = msg.GetValue()
	_ = msg.GetHeaders()
	_ = msg.GetTopic()
	_ = msg.GetTimestamp()
}

func TestProducerInterface_Compliance(t *testing.T) {
	// 确保 *KafkaProducer 实现了 Producer 接口
	var _ Producer = &KafkaProducer{}
}

func TestConsumerInterface_Compliance(t *testing.T) {
	// 确保 *KafkaConsumer 实现了 Consumer 接口
	var _ Consumer = &KafkaConsumer{}
}

func TestMQFactoryInterface_Compliance(t *testing.T) {
	// 确保 *KafkaFactory 实现了 MQFactory 接口
	var _ MQFactory = &KafkaFactory{}

	// 确保 *RabbitMQFactory 实现了 MQFactory 接口
	var _ MQFactory = &RabbitMQFactory{}
}

// ===== Factory Tests =====

func TestKafkaFactory_GetMQType(t *testing.T) {
	factory := NewKafkaFactory()

	if factory.GetMQType() != MQTypeKafka {
		t.Errorf("GetMQType 期望 Kafka, 得到 %s", factory.GetMQType())
	}
}

func TestKafkaFactory_CreateProducer(t *testing.T) {
	factory := NewKafkaFactory()

	producer, err := factory.CreateProducer()
	if err != nil {
		t.Errorf("CreateProducer 不应该返回错误: %v", err)
	}

	if producer == nil {
		t.Error("Producer 不应该为 nil")
	}
}

func TestKafkaFactory_CreateConsumer(t *testing.T) {
	factory := NewKafkaFactory()

	consumer, err := factory.CreateConsumer()
	if err != nil {
		t.Errorf("CreateConsumer 不应该返回错误: %v", err)
	}

	if consumer == nil {
		t.Error("Consumer 不应该为 nil")
	}
}

func TestRabbitMQFactory_GetMQType(t *testing.T) {
	factory := NewRabbitMQFactory()

	if factory.GetMQType() != MQTypeRabbitMQ {
		t.Errorf("GetMQType 期望 RabbitMQ, 得到 %s", factory.GetMQType())
	}
}

func TestRabbitMQFactory_CreateProducer(t *testing.T) {
	factory := NewRabbitMQFactory()

	// 这里会因为无法连接 RabbitMQ 而返回错误
	producer, err := factory.CreateProducer()
	if err == nil && producer != nil {
		t.Log("RabbitMQ 生产者创建成功（可能连接到实际的 RabbitMQ）")
	}
}

func TestRabbitMQFactory_CreateConsumer(t *testing.T) {
	factory := NewRabbitMQFactory()

	// 这里会因为无法连接 RabbitMQ 而返回错误
	consumer, err := factory.CreateConsumer()
	if err == nil && consumer != nil {
		t.Log("RabbitMQ 消费者创建成功（可能连接到实际的 RabbitMQ）")
	}
}

// ===== Kafka Producer Options Tests =====

func TestKafkaProducerOptions_Defaults(t *testing.T) {
	opts := &KafkaProducerOptions{
		Brokers:      []string{"localhost:9092"},
		Topic:        "default",
		BatchSize:    10,
		BatchTimeout: 10 * time.Millisecond,
	}

	if opts.Topic != "default" {
		t.Errorf("Topic 期望 'default', 得到 %s", opts.Topic)
	}

	if opts.BatchSize != 10 {
		t.Errorf("BatchSize 期望 10, 得到 %d", opts.BatchSize)
	}
}

func TestWithKafkaBrokers(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaBrokers("broker1:9092", "broker2:9092")(opts)

	if len(opts.Brokers) != 2 {
		t.Errorf("Brokers 长度期望 2, 得到 %d", len(opts.Brokers))
	}

	if opts.Brokers[0] != "broker1:9092" {
		t.Errorf("Brokers[0] 期望 'broker1:9092', 得到 %s", opts.Brokers[0])
	}
}

func TestWithKafkaTopic(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaTopic("my-topic")(opts)

	if opts.Topic != "my-topic" {
		t.Errorf("Topic 期望 'my-topic', 得到 %s", opts.Topic)
	}
}

func TestWithKafkaBatchSize(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaBatchSize(100)(opts)

	if opts.BatchSize != 100 {
		t.Errorf("BatchSize 期望 100, 得到 %d", opts.BatchSize)
	}
}

func TestWithKafkaBatchTimeout(t *testing.T) {
	opts := &KafkaProducerOptions{}

	timeout := 5 * time.Second
	WithKafkaBatchTimeout(timeout)(opts)

	if opts.BatchTimeout != timeout {
		t.Errorf("BatchTimeout 期望 %v, 得到 %v", timeout, opts.BatchTimeout)
	}
}

func TestWithKafkaAsync(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaAsync(true)(opts)

	if !opts.Async {
		t.Error("Async 应该为 true")
	}
}

func TestWithKafkaMaxAttempts(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaMaxAttempts(5)(opts)

	if opts.MaxAttempts != 5 {
		t.Errorf("MaxAttempts 期望 5, 得到 %d", opts.MaxAttempts)
	}
}

func TestKafkaProducerOptions_AllOptions(t *testing.T) {
	opts := &KafkaProducerOptions{}

	WithKafkaBrokers("broker1:9092")(opts)
	WithKafkaTopic("test-topic")(opts)
	WithKafkaBatchSize(50)(opts)
	WithKafkaBatchTimeout(time.Second)(opts)
	WithKafkaAsync(false)(opts)
	WithKafkaMaxAttempts(3)(opts)

	if opts.Topic != "test-topic" {
		t.Error("Topic 设置失败")
	}

	if opts.BatchSize != 50 {
		t.Error("BatchSize 设置失败")
	}

	if opts.Async {
		t.Error("Async 设置失败")
	}
}

// ===== Kafka Consumer Options Tests =====

func TestKafkaConsumerOptions_Defaults(t *testing.T) {
	opts := &KafkaConsumerOptions{
		Brokers:     []string{"localhost:9092"},
		Topic:       "default",
		GroupID:     "default-group",
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxAttempts: 3,
	}

	if opts.Topic != "default" {
		t.Errorf("Topic 期望 'default', 得到 %s", opts.Topic)
	}

	if opts.GroupID != "default-group" {
		t.Errorf("GroupID 期望 'default-group', 得到 %s", opts.GroupID)
	}
}

func TestWithKafkaConsumerBrokers(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerBrokers("kafka1:9092", "kafka2:9092")(opts)

	if len(opts.Brokers) != 2 {
		t.Errorf("Brokers 长度期望 2, 得到 %d", len(opts.Brokers))
	}
}

func TestWithKafkaConsumerTopic(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerTopic("consumer-topic")(opts)

	if opts.Topic != "consumer-topic" {
		t.Errorf("Topic 期望 'consumer-topic', 得到 %s", opts.Topic)
	}
}

func TestWithKafkaConsumerGroupID(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerGroupID("my-group")(opts)

	if opts.GroupID != "my-group" {
		t.Errorf("GroupID 期望 'my-group', 得到 %s", opts.GroupID)
	}
}

func TestWithKafkaConsumerMinBytes(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerMinBytes(1024)(opts)

	if opts.MinBytes != 1024 {
		t.Errorf("MinBytes 期望 1024, 得到 %d", opts.MinBytes)
	}
}

func TestWithKafkaConsumerMaxBytes(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerMaxBytes(5e6)(opts)

	if opts.MaxBytes != 5e6 {
		t.Errorf("MaxBytes 期望 5e6, 得到 %d", opts.MaxBytes)
	}
}

func TestWithKafkaConsumerMaxAttempts(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerMaxAttempts(10)(opts)

	if opts.MaxAttempts != 10 {
		t.Errorf("MaxAttempts 期望 10, 得到 %d", opts.MaxAttempts)
	}
}

func TestWithKafkaConsumerStartOffset(t *testing.T) {
	opts := &KafkaConsumerOptions{}

	WithKafkaConsumerStartOffset(100)(opts)

	if opts.StartOffset != 100 {
		t.Errorf("StartOffset 期望 100, 得到 %d", opts.StartOffset)
	}
}

// ===== RabbitMQ Producer Options Tests =====

func TestRabbitMQProducerOptions_Defaults(t *testing.T) {
	opts := &RabbitMQProducerOptions{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "default",
		ExchangeType: "direct",
		RoutingKey:   "default",
		ContentType:  "application/json",
	}

	if opts.Exchange != "default" {
		t.Errorf("Exchange 期望 'default', 得到 %s", opts.Exchange)
	}

	if opts.ExchangeType != "direct" {
		t.Errorf("ExchangeType 期望 'direct', 得到 %s", opts.ExchangeType)
	}
}

func TestWithRabbitMQURL(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQURL("amqp://user:pass@rabbitmq:5672/")(opts)

	if opts.URL != "amqp://user:pass@rabbitmq:5672/" {
		t.Errorf("URL 设置失败")
	}
}

func TestWithRabbitMQExchange(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQExchange("my-exchange")(opts)

	if opts.Exchange != "my-exchange" {
		t.Errorf("Exchange 期望 'my-exchange', 得到 %s", opts.Exchange)
	}
}

func TestWithRabbitMQExchangeType(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQExchangeType("topic")(opts)

	if opts.ExchangeType != "topic" {
		t.Errorf("ExchangeType 期望 'topic', 得到 %s", opts.ExchangeType)
	}
}

func TestWithRabbitMQRoutingKey(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQRoutingKey("my-routing-key")(opts)

	if opts.RoutingKey != "my-routing-key" {
		t.Errorf("RoutingKey 期望 'my-routing-key', 得到 %s", opts.RoutingKey)
	}
}

func TestWithRabbitMQContentType(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQContentType("text/plain")(opts)

	if opts.ContentType != "text/plain" {
		t.Errorf("ContentType 期望 'text/plain', 得到 %s", opts.ContentType)
	}
}

func TestWithRabbitMQPriority(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQPriority(5)(opts)

	if opts.Priority != 5 {
		t.Errorf("Priority 期望 5, 得到 %d", opts.Priority)
	}
}

func TestWithRabbitMQMandatory(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQMandatory(true)(opts)

	if !opts.Mandatory {
		t.Error("Mandatory 应该为 true")
	}
}

func TestWithRabbitMQImmediate(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQImmediate(true)(opts)

	if !opts.Immediate {
		t.Error("Immediate 应该为 true")
	}
}

func TestRabbitMQProducerOptions_AllOptions(t *testing.T) {
	opts := &RabbitMQProducerOptions{}

	WithRabbitMQURL("amqp://user:pass@localhost:5672/")(opts)
	WithRabbitMQExchange("test-exchange")(opts)
	WithRabbitMQExchangeType("fanout")(opts)
	WithRabbitMQRoutingKey("test-key")(opts)
	WithRabbitMQContentType("application/octet-stream")(opts)
	WithRabbitMQPriority(10)(opts)
	WithRabbitMQMandatory(false)(opts)
	WithRabbitMQImmediate(false)(opts)

	if opts.ExchangeType != "fanout" {
		t.Error("ExchangeType 设置失败")
	}

	if opts.Priority != 10 {
		t.Error("Priority 设置失败")
	}
}

// ===== RabbitMQ Consumer Options Tests =====

func TestRabbitMQConsumerOptions_Defaults(t *testing.T) {
	opts := &RabbitMQConsumerOptions{
		URL:           "amqp://guest:guest@localhost:5672/",
		Exchange:      "default",
		ExchangeType:  "direct",
		Queue:         "default",
		RoutingKey:    "default",
		AutoAck:       false,
		PrefetchCount: 10,
	}

	if opts.Queue != "default" {
		t.Errorf("Queue 期望 'default', 得到 %s", opts.Queue)
	}

	if opts.AutoAck {
		t.Error("AutoAck 默认应该为 false")
	}

	if opts.PrefetchCount != 10 {
		t.Errorf("PrefetchCount 期望 10, 得到 %d", opts.PrefetchCount)
	}
}

func TestWithRabbitMQConsumerURL(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerURL("amqp://user:pass@rabbitmq:5672/")(opts)

	if opts.URL != "amqp://user:pass@rabbitmq:5672/" {
		t.Errorf("URL 设置失败")
	}
}

func TestWithRabbitMQConsumerExchange(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerExchange("consumer-exchange")(opts)

	if opts.Exchange != "consumer-exchange" {
		t.Errorf("Exchange 期望 'consumer-exchange', 得到 %s", opts.Exchange)
	}
}

func TestWithRabbitMQConsumerExchangeType(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerExchangeType("headers")(opts)

	if opts.ExchangeType != "headers" {
		t.Errorf("ExchangeType 期望 'headers', 得到 %s", opts.ExchangeType)
	}
}

func TestWithRabbitMQConsumerQueue(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerQueue("my-queue")(opts)

	if opts.Queue != "my-queue" {
		t.Errorf("Queue 期望 'my-queue', 得到 %s", opts.Queue)
	}
}

func TestWithRabbitMQConsumerRoutingKey(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerRoutingKey("consumer-key")(opts)

	if opts.RoutingKey != "consumer-key" {
		t.Errorf("RoutingKey 期望 'consumer-key', 得到 %s", opts.RoutingKey)
	}
}

func TestWithRabbitMQConsumerTag(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerTag("my-consumer-tag")(opts)

	if opts.ConsumerTag != "my-consumer-tag" {
		t.Errorf("ConsumerTag 期望 'my-consumer-tag', 得到 %s", opts.ConsumerTag)
	}
}

func TestWithRabbitMQConsumerAutoAck(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerAutoAck(true)(opts)

	if !opts.AutoAck {
		t.Error("AutoAck 应该为 true")
	}
}

func TestWithRabbitMQConsumerPrefetchCount(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerPrefetchCount(50)(opts)

	if opts.PrefetchCount != 50 {
		t.Errorf("PrefetchCount 期望 50, 得到 %d", opts.PrefetchCount)
	}
}

func TestWithRabbitMQConsumerExclusive(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerExclusive(true)(opts)

	if !opts.Exclusive {
		t.Error("Exclusive 应该为 true")
	}
}

func TestRabbitMQConsumerOptions_AllOptions(t *testing.T) {
	opts := &RabbitMQConsumerOptions{}

	WithRabbitMQConsumerURL("amqp://user:pass@localhost:5672/")(opts)
	WithRabbitMQConsumerExchange("test-exchange")(opts)
	WithRabbitMQConsumerExchangeType("direct")(opts)
	WithRabbitMQConsumerQueue("test-queue")(opts)
	WithRabbitMQConsumerRoutingKey("test-routing")(opts)
	WithRabbitMQConsumerTag("test-tag")(opts)
	WithRabbitMQConsumerAutoAck(true)(opts)
	WithRabbitMQConsumerPrefetchCount(20)(opts)
	WithRabbitMQConsumerExclusive(true)(opts)

	if opts.Queue != "test-queue" {
		t.Error("Queue 设置失败")
	}

	if !opts.AutoAck {
		t.Error("AutoAck 设置失败")
	}

	if opts.PrefetchCount != 20 {
		t.Error("PrefetchCount 设置失败")
	}
}

// ===== Kafka Consumed Message Tests =====

func TestKafkaConsumedMessage(t *testing.T) {
	msg := &KafkaConsumedMessage{
		key:       "test-key",
		value:     []byte("test-value"),
		headers:   map[string]interface{}{"h1": "v1"},
		topic:     "test-topic",
		timestamp: 1234567890,
	}

	if msg.GetKey() != "test-key" {
		t.Errorf("Key 期望 'test-key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "test-value" {
		t.Errorf("Value 期望 'test-value', 得到 %s", string(msg.GetValue()))
	}

	if msg.GetTopic() != "test-topic" {
		t.Errorf("Topic 期望 'test-topic', 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != 1234567890 {
		t.Errorf("Timestamp 期望 1234567890, 得到 %d", msg.GetTimestamp())
	}

	if msg.GetHeaders()["h1"] != "v1" {
		t.Error("Headers 设置失败")
	}
}

// ===== RabbitMQ Consumed Message Tests =====

func TestRabbitMQConsumedMessage(t *testing.T) {
	msg := &RabbitMQConsumedMessage{
		routingKey: "routing-key",
		body:       []byte("body-content"),
		headers:    map[string]interface{}{"h2": "v2"},
		topic:      "exchange-name",
		timestamp:  9876543210,
	}

	if msg.GetKey() != "routing-key" {
		t.Errorf("Key 期望 'routing-key', 得到 %s", msg.GetKey())
	}

	if string(msg.GetValue()) != "body-content" {
		t.Errorf("Value 期望 'body-content', 得到 %s", string(msg.GetValue()))
	}

	if msg.GetTopic() != "exchange-name" {
		t.Errorf("Topic 期望 'exchange-name', 得到 %s", msg.GetTopic())
	}

	if msg.GetTimestamp() != 9876543210 {
		t.Errorf("Timestamp 期望 9876543210, 得到 %d", msg.GetTimestamp())
	}

	if msg.GetHeaders()["h2"] != "v2" {
		t.Error("Headers 设置失败")
	}
}

// ===== Factory Manager Tests =====

func TestFactoryManager_New(t *testing.T) {
	manager := NewFactoryManager()

	if manager == nil {
		t.Error("FactoryManager 不应该为 nil")
	}

	if manager.factories == nil {
		t.Error("FactoryManager.factories 不应该为 nil")
	}
}

func TestFactoryManager_Register(t *testing.T) {
	manager := NewFactoryManager()

	manager.Register(NewKafkaFactory())
	manager.Register(NewRabbitMQFactory())

	if len(manager.factories) != 2 {
		t.Errorf("factories 数量期望 2, 得到 %d", len(manager.factories))
	}
}

func TestFactoryManager_GetFactory(t *testing.T) {
	manager := NewFactoryManager()
	manager.Register(NewKafkaFactory())
	manager.Register(NewRabbitMQFactory())

	factory, err := manager.GetFactory(MQTypeKafka)
	if err != nil {
		t.Errorf("GetFactory 不应该返回错误: %v", err)
	}
	if factory.GetMQType() != MQTypeKafka {
		t.Errorf("Factory 类型期望 Kafka, 得到 %s", factory.GetMQType())
	}

	factory, err = manager.GetFactory(MQTypeRabbitMQ)
	if err != nil {
		t.Errorf("GetFactory 不应该返回错误: %v", err)
	}
	if factory.GetMQType() != MQTypeRabbitMQ {
		t.Errorf("Factory 类型期望 RabbitMQ, 得到 %s", factory.GetMQType())
	}

	_, err = manager.GetFactory(MQTypeRocketMQ)
	if err == nil {
		t.Error("GetFactory 应该对未注册的 MQ 类型返回错误")
	}
}

func TestFactoryManager_CreateProducer(t *testing.T) {
	manager := NewFactoryManager()
	manager.Register(NewKafkaFactory())

	producer, err := manager.CreateProducer(MQTypeKafka)
	if err != nil {
		t.Errorf("CreateProducer 不应该返回错误: %v", err)
	}

	if producer == nil {
		t.Error("Producer 不应该为 nil")
	}

	_, err = manager.CreateProducer(MQTypeRocketMQ)
	if err == nil {
		t.Error("CreateProducer 应该对未注册的 MQ 类型返回错误")
	}
}

func TestFactoryManager_CreateConsumer(t *testing.T) {
	manager := NewFactoryManager()
	manager.Register(NewKafkaFactory())

	consumer, err := manager.CreateConsumer(MQTypeKafka)
	if err != nil {
		t.Errorf("CreateConsumer 不应该返回错误: %v", err)
	}

	if consumer == nil {
		t.Error("Consumer 不应该为 nil")
	}

	_, err = manager.CreateConsumer(MQTypeRocketMQ)
	if err == nil {
		t.Error("CreateConsumer 应该对未注册的 MQ 类型返回错误")
	}
}

// ===== InitMQ and GetMQManager Tests =====

func TestInitMQ(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	manager := InitMQ()

	if manager == nil {
		t.Error("InitMQ 不应该返回 nil")
	}

	if len(manager.factories) != 2 {
		t.Errorf("InitMQ 应该注册 2 个工厂, 得到 %d", len(manager.factories))
	}
}

func TestGetMQManager(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	manager := GetMQManager()

	if manager == nil {
		t.Error("GetMQManager 不应该返回 nil")
	}

	// 再次调用应该返回同一个实例
	manager2 := GetMQManager()
	if manager != manager2 {
		t.Error("GetMQManager 应该返回同一个实例")
	}
}

// ===== NewProducer and NewConsumer Tests =====

func TestNewProducer_Kafka(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	producer, err := NewProducer(
		MQTypeKafka,
		WithMQType(MQTypeKafka),
	)

	if err != nil {
		t.Errorf("NewProducer 不应该返回错误: %v", err)
	}

	if producer == nil {
		t.Error("Producer 不应该为 nil")
	}

	// 清理
	if closer, ok := producer.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func TestNewConsumer_Kafka(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	consumer, err := NewConsumer(
		MQTypeKafka,
		WithConsumerGroupID("test-group"),
	)

	if err != nil {
		t.Errorf("NewConsumer 不应该返回错误: %v", err)
	}

	if consumer == nil {
		t.Error("Consumer 不应该为 nil")
	}

	// 清理
	_ = consumer.Close()
}

// ===== Common Options Tests =====

func TestCommonProducerOptions(t *testing.T) {
	opts := &CommonOptions{
		MQType:  MQTypeKafka,
		Brokers: []string{"localhost:9092"},
		Topic:   "common-topic",
	}

	if opts.MQType != MQTypeKafka {
		t.Errorf("MQType 期望 Kafka, 得到 %s", opts.MQType)
	}

	if opts.Topic != "common-topic" {
		t.Errorf("Topic 期望 'common-topic', 得到 %s", opts.Topic)
	}
}

func TestCommonConsumerOptions(t *testing.T) {
	opts := &CommonConsumerOptions{
		MQType:  MQTypeRabbitMQ,
		Brokers: []string{"amqp://localhost:5672"},
		Topic:   "common-consumer-topic",
		GroupID: "common-group",
	}

	if opts.MQType != MQTypeRabbitMQ {
		t.Errorf("MQType 期望 RabbitMQ, 得到 %s", opts.MQType)
	}

	if opts.GroupID != "common-group" {
		t.Errorf("GroupID 期望 'common-group', 得到 %s", opts.GroupID)
	}
}

func TestCreateProducerWithCommonOpts_Kafka(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	producer, err := CreateProducerWithCommonOpts(
		WithCommonMQType(MQTypeKafka),
		WithCommonBrokers("broker1:9092", "broker2:9092"),
		WithCommonTopic("common-test-topic"),
	)

	if err != nil {
		t.Errorf("CreateProducerWithCommonOpts 不应该返回错误: %v", err)
	}

	if producer == nil {
		t.Error("Producer 不应该为 nil")
	}

	// 清理
	if closer, ok := producer.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func TestCreateConsumerWithCommonOpts_Kafka(t *testing.T) {
	// 重置全局管理器
	defaultManager = nil

	consumer, err := CreateConsumerWithCommonOpts(
		WithCommonConsumerMQType(MQTypeKafka),
		WithCommonConsumerBrokers("broker1:9092"),
		WithCommonConsumerTopic("common-consumer-topic"),
		WithCommonConsumerGroupID("common-consumer-group"),
	)

	if err != nil {
		t.Errorf("CreateConsumerWithCommonOpts 不应该返回错误: %v", err)
	}

	if consumer == nil {
		t.Error("Consumer 不应该为 nil")
	}

	// 清理
	_ = consumer.Close()
}

func TestCreateProducerWithCommonOpts_Unsupported(t *testing.T) {
	producer, err := CreateProducerWithCommonOpts(
		WithCommonMQType(MQTypeRocketMQ),
	)

	if err == nil {
		t.Error("CreateProducerWithCommonOpts 应该对不支持的 MQ 类型返回错误")
	}

	if producer != nil {
		t.Error("Producer 应该为 nil")
	}
}

func TestCreateConsumerWithCommonOpts_Unsupported(t *testing.T) {
	consumer, err := CreateConsumerWithCommonOpts(
		WithCommonConsumerMQType(MQTypeRocketMQ),
	)

	if err == nil {
		t.Error("CreateConsumerWithCommonOpts 应该对不支持的 MQ 类型返回错误")
	}

	if consumer != nil {
		t.Error("Consumer 应该为 nil")
	}
}

// ===== Context Tests =====

func TestProducer_Send_WithoutConnection(t *testing.T) {
	producer := &KafkaProducer{}
	ctx := context.Background()
	msg := NewBaseMessage("key", []byte("value"))

	// 在没有实际连接的情况下，Send 应该会失败
	err := producer.Send(ctx, msg)
	if err == nil {
		t.Error("Send 应该返回错误（没有实际连接）")
	}
}

func TestProducer_SendJSON_WithoutConnection(t *testing.T) {
	producer := &KafkaProducer{}
	ctx := context.Background()

	// 在没有实际连接的情况下，SendJSON 应该会失败
	err := producer.SendJSON(ctx, "topic", "key", map[string]string{"test": "value"})
	if err == nil {
		t.Error("SendJSON 应该返回错误（没有实际连接）")
	}
}

func TestProducer_SendBatch_WithoutConnection(t *testing.T) {
	producer := &KafkaProducer{}
	ctx := context.Background()
	msgs := []Message{
		NewBaseMessage("key1", []byte("value1")),
		NewBaseMessage("key2", []byte("value2")),
	}

	// 在没有实际连接的情况下，SendBatch 应该会失败
	err := producer.SendBatch(ctx, msgs)
	if err == nil {
		t.Error("SendBatch 应该返回错误（没有实际连接）")
	}
}

func TestProducer_SendBatch_Empty(t *testing.T) {
	producer := &KafkaProducer{}
	ctx := context.Background()

	// 空的批量消息应该成功
	err := producer.SendBatch(ctx, []Message{})
	if err != nil {
		t.Errorf("SendBatch 空消息不应该返回错误: %v", err)
	}
}

func TestConsumer_Subscribe_WithoutConnection(t *testing.T) {
	consumer := &KafkaConsumer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	handler := func(msg Message) error {
		return nil
	}

	// 立即取消的上下文应该立即返回
	err := consumer.Subscribe(ctx, handler)
	if err == nil {
		t.Error("Subscribe 应该返回错误（上下文已取消）")
	}
}

func TestConsumer_SubscribeTopic_WithoutConnection(t *testing.T) {
	consumer := &KafkaConsumer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := func(msg Message) error {
		return nil
	}

	err := consumer.SubscribeTopic(ctx, "topic", handler)
	if err == nil {
		t.Error("SubscribeTopic 应该返回错误（上下文已取消）")
	}
}
