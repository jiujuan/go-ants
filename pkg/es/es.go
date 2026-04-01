package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/jiujuan/go-ants/pkg/log"
)

// Client Elasticsearch 客户端
type Client struct {
	client *elasticsearch.Client
	opts   *Options
}

// Option 是选项函数
type Option func(*Options)

type Options struct {
	Addresses    []string
	Username     string
	Password     string
	APIKey       string
	CloudID      string
	IndexPrefix  string
	RetryOnError func(*esapi.Request, *esapi.Response, error) (bool, error)
	Transport    http.RoundTripper
}

// WithESAddresses 设置地址列表
func WithESAddresses(addresses ...string) Option {
	return func(o *Options) {
		o.Addresses = addresses
	}
}

// WithESUsername 设置用户名
func WithESUsername(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// WithESPassword 设置密码
func WithESPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithESAPIKey 设置 API Key
func WithESAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithESCloudID 设置 Cloud ID
func WithESCloudID(cloudID string) Option {
	return func(o *Options) {
		o.CloudID = cloudID
	}
}

// WithESIndexPrefix 设置索引前缀
func WithESIndexPrefix(prefix string) Option {
	return func(o *Options) {
		o.IndexPrefix = prefix
	}
}

// New 创建新的 Elasticsearch 客户端
func New(opts ...Option) (*Client, error) {
	options := &Options{
		Addresses:   []string{"http://localhost:9200"},
		IndexPrefix: "go-ants",
	}

	for _, opt := range opts {
		opt(options)
	}

	cfg := elasticsearch.Config{
		Addresses:     options.Addresses,
		Username:      options.Username,
		Password:      options.Password,
		APIKey:        options.APIKey,
		CloudID:       options.CloudID,
		RetryOnStatus: []int{502, 503, 504, 429},
		MaxRetries:    3,
		RetryBackoff:  func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	log.Info("elasticsearch connected",
		log.String("addresses", strings.Join(options.Addresses, ",")))

	return &Client{
		client: client,
		opts:   options,
	}, nil
}

// ===== 索引操作 =====

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, index string, mapping string) error {
	res, err := c.client.Indices.Create(
		index,
		c.client.Indices.Create.WithContext(ctx),
		c.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}

	return nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	res, err := c.client.Indices.Delete(
		[]string{index},
		c.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.String())
	}

	return nil
}

// ExistsIndex 检查索引是否存在
func (c *Client) ExistsIndex(ctx context.Context, index string) (bool, error) {
	res, err := c.client.Indices.Exists(
		[]string{index},
		c.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check index exists: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// ===== 文档操作 =====

// Index 索引文档
func (c *Client) Index(ctx context.Context, index, id string, document interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(document); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	res, err := c.client.Index(
		index,
		&buf,
		c.client.Index.WithContext(ctx),
		c.client.Index.WithDocumentID(id),
		c.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index document error: %s", res.String())
	}

	return nil
}

// Get 获取文档
func (c *Client) Get(ctx context.Context, index, id string) ([]byte, error) {
	res, err := c.client.Get(
		index,
		id,
		c.client.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("get document error: %s", res.String())
	}

	return res.Body.Bytes(), nil
}

// Delete 删除文档
func (c *Client) Delete(ctx context.Context, index, id string) error {
	res, err := c.client.Delete(
		index,
		id,
		c.client.Delete.WithContext(ctx),
		c.client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete document error: %s", res.String())
	}

	return nil
}

// Update 更新文档
func (c *Client) Update(ctx context.Context, index, id string, document interface{}) error {
	var buf bytes.Buffer
	update := map[string]interface{}{
		"doc": document,
	}
	if err := json.NewEncoder(&buf).Encode(update); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	res, err := c.client.Update(
		index,
		id,
		&buf,
		c.client.Update.WithContext(ctx),
		c.client.Update.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("update document error: %s", res.String())
	}

	return nil
}

// ===== 批量操作 =====

// Bulk 批量操作
func (c *Client) Bulk(ctx context.Context, operations []BulkOperation) error {
	var buf bytes.Buffer
	for _, op := range operations {
		meta, err := json.Marshal(op.Meta)
		if err != nil {
			return fmt.Errorf("failed to encode bulk meta: %w", err)
		}
		buf.WriteString(string(meta) + "\n")

		if op.Document != nil {
			doc, err := json.Marshal(op.Document)
			if err != nil {
				return fmt.Errorf("failed to encode bulk document: %w", err)
			}
			buf.WriteString(string(doc) + "\n")
		}
	}

	res, err := c.client.Bulk(
		&buf,
		c.client.Bulk.WithContext(ctx),
		c.client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to execute bulk: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.String())
	}

	return nil
}

// BulkOperation 批量操作
type BulkOperation struct {
	Meta     BulkMeta
	Document interface{}
}

// BulkMeta 批量操作元数据
type BulkMeta struct {
	Index  string `json:"_index,omitempty"`
	ID     string `json:"_id,omitempty"`
	Type   string `json:"_type,omitempty"`
	Action string `json:"-"`
}

// ===== 搜索操作 =====

// Search 搜索
func (c *Client) Search(ctx context.Context, req SearchRequest) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Query); err != nil {
		return nil, fmt.Errorf("failed to encode search query: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(req.Index...),
		c.client.Search.WithBody(&buf),
		c.client.Search.WithTrackTotalHits(req.TrackTotalHits),
		c.client.Search.WithFrom(req.From),
		c.client.Search.WithSize(req.Size),
		c.client.Search.WithSort(req.Sort...),
		c.client.Search.WithSource(req.Source),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	return res.Body.Bytes(), nil
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Index          []string
	Query          interface{}
	From           int
	Size           int
	Sort           []string
	Source         []string
	TrackTotalHits bool
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Hits     SearchHits
	Total    int64
	MaxScore float64
}

// SearchHits 搜索命中
type SearchHits struct {
	Hits  []SearchHit
	Total int64
}

// SearchHit 单个命中
type SearchHit struct {
	ID     string          `json:"_id"`
	Index  string          `json:"_index"`
	Score  float64         `json:"_score"`
	Source json.RawMessage `json:"_source"`
}

// ===== 聚合操作 =====

// Aggregate 聚合
func (c *Client) Aggregate(ctx context.Context, req AggregateRequest) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Query); err != nil {
		return nil, fmt.Errorf("failed to encode aggregate query: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(req.Index...),
		c.client.Search.WithBody(&buf),
		c.client.Search.WithSize(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("aggregate error: %s", res.String())
	}

	return res.Body.Bytes(), nil
}

// AggregateRequest 聚合请求
type AggregateRequest struct {
	Index []string
	Query interface{}
	Aggs  map[string]map[string]interface{}
	Size  int
}

// ===== 工具函数 =====

// GetClient 获取底层客户端
func (c *Client) GetClient() *elasticsearch.Client {
	return c.client
}

// Ping ping Elasticsearch
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.client.Ping(c.client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping error: %s", res.String())
	}

	return nil
}

// 获取索引名称（带前缀）
func (c *Client) GetIndexName(name string) string {
	if c.opts.IndexPrefix != "" {
		return c.opts.IndexPrefix + "-" + name
	}
	return name
}
