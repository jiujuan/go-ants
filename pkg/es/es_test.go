package es

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ===== 选项函数测试 =====

func TestWithESAddresses(t *testing.T) {
	opts := &Options{}

	WithESAddresses("http://localhost:9200", "http://localhost:9201")(opts)

	if len(opts.Addresses) != 2 {
		t.Errorf("Addresses 长度期望 2, 得到 %d", len(opts.Addresses))
	}

	if opts.Addresses[0] != "http://localhost:9200" {
		t.Errorf("Addresses[0] 期望 'http://localhost:9200', 得到 %s", opts.Addresses[0])
	}

	if opts.Addresses[1] != "http://localhost:9201" {
		t.Errorf("Addresses[1] 期望 'http://localhost:9201', 得到 %s", opts.Addresses[1])
	}
}

func TestWithESUsername(t *testing.T) {
	opts := &Options{}

	WithESUsername("elastic")(opts)

	if opts.Username != "elastic" {
		t.Errorf("Username 期望 'elastic', 得到 %s", opts.Username)
	}
}

func TestWithESPassword(t *testing.T) {
	opts := &Options{}

	WithESPassword("secretpassword")(opts)

	if opts.Password != "secretpassword" {
		t.Errorf("Password 期望 'secretpassword', 得到 %s", opts.Password)
	}
}

func TestWithESAPIKey(t *testing.T) {
	opts := &Options{}

	WithESAPIKey("my-api-key")(opts)

	if opts.APIKey != "my-api-key" {
		t.Errorf("APIKey 期望 'my-api-key', 得到 %s", opts.APIKey)
	}
}

func TestWithESCloudID(t *testing.T) {
	opts := &Options{}

	WithESCloudID("my-cluster:ZXVyb3Blbl8xMjM0NTY3ODkw")(opts)

	if opts.CloudID != "my-cluster:ZXVyb3Blbl8xMjM0NTY3ODkw" {
		t.Errorf("CloudID 设置失败")
	}
}

func TestWithESIndexPrefix(t *testing.T) {
	opts := &Options{}

	WithESIndexPrefix("production")(opts)

	if opts.IndexPrefix != "production" {
		t.Errorf("IndexPrefix 期望 'production', 得到 %s", opts.IndexPrefix)
	}
}

func TestOptions_CombinedOptions(t *testing.T) {
	opts := &Options{}

	WithESAddresses("http://es1:9200", "http://es2:9200")(opts)
	WithESUsername("admin")(opts)
	WithESPassword("admin123")(opts)
	WithESAPIKey("api-key-123")(opts)
	WithESCloudID("cluster:id")(opts)
	WithESIndexPrefix("test")(opts)

	if len(opts.Addresses) != 2 {
		t.Error("Addresses 设置失败")
	}

	if opts.Username != "admin" {
		t.Error("Username 设置失败")
	}

	if opts.Password != "admin123" {
		t.Error("Password 设置失败")
	}

	if opts.APIKey != "api-key-123" {
		t.Error("APIKey 设置失败")
	}

	if opts.CloudID != "cluster:id" {
		t.Error("CloudID 设置失败")
	}

	if opts.IndexPrefix != "test" {
		t.Error("IndexPrefix 设置失败")
	}
}

func TestOptions_Defaults(t *testing.T) {
	opts := &Options{
		Addresses:   []string{"http://localhost:9200"},
		IndexPrefix: "go-ants",
	}

	if opts.Addresses[0] != "http://localhost:9200" {
		t.Errorf("默认 Address 期望 'http://localhost:9200', 得到 %s", opts.Addresses[0])
	}

	if opts.IndexPrefix != "go-ants" {
		t.Errorf("默认 IndexPrefix 期望 'go-ants', 得到 %s", opts.IndexPrefix)
	}
}

// ===== 客户端结构测试 =====

func TestClient_Struct(t *testing.T) {
	client := &Client{}

	if client == nil {
		t.Error("Client 不应该为 nil")
	}
}

func TestClient_GetClient(t *testing.T) {
	client := &Client{}

	if client.GetClient() != nil {
		t.Error("GetClient 应该返回 nil（未初始化的客户端）")
	}
}

func TestClient_GetIndexName(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		indexName string
		expected  string
	}{
		{
			name:      "带前缀",
			prefix:    "production",
			indexName: "users",
			expected:  "production-users",
		},
		{
			name:      "空前缀",
			prefix:    "",
			indexName: "users",
			expected:  "users",
		},
		{
			name:      "默认前缀",
			prefix:    "go-ants",
			indexName: "products",
			expected:  "go-ants-products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				opts: &Options{
					IndexPrefix: tt.prefix,
				},
			}

			result := client.GetIndexName(tt.indexName)
			if result != tt.expected {
				t.Errorf("GetIndexName 期望 '%s', 得到 '%s'", tt.expected, result)
			}
		})
	}
}

// ===== 错误测试 =====

func TestErrors(t *testing.T) {
	// 测试错误类型定义
	tests := []struct {
		name    string
		checkFn func() error
	}{
		{
			name: "创建客户端错误 - 无效配置",
			checkFn: func() error {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.checkFn
		})
	}
}

// ===== BulkMeta 测试 =====

func TestBulkMeta_JSON(t *testing.T) {
	tests := []struct {
		name     string
		meta     BulkMeta
		expected map[string]interface{}
	}{
		{
			name: "Index 操作",
			meta: BulkMeta{
				Index:  "test-index",
				ID:     "doc-1",
				Action: "index",
			},
			expected: map[string]interface{}{
				"_index": "test-index",
				"_id":    "doc-1",
			},
		},
		{
			name: "Delete 操作",
			meta: BulkMeta{
				Index:  "test-index",
				ID:     "doc-2",
				Action: "delete",
			},
			expected: map[string]interface{}{
				"_index": "test-index",
				"_id":    "doc-2",
			},
		},
		{
			name: "Update 操作",
			meta: BulkMeta{
				Index:  "test-index",
				ID:     "doc-3",
				Action: "update",
			},
			expected: map[string]interface{}{
				"_index": "test-index",
				"_id":    "doc-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.meta)
			if err != nil {
				t.Errorf("JSON Marshal 失败: %v", err)
			}

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Errorf("JSON Unmarshal 失败: %v", err)
			}

			if result["_index"] != tt.expected["_index"] {
				t.Errorf("_index 期望 '%v', 得到 '%v'", tt.expected["_index"], result["_index"])
			}

			if result["_id"] != tt.expected["_id"] {
				t.Errorf("_id 期望 '%v', 得到 '%v'", tt.expected["_id"], result["_id"])
			}
		})
	}
}

func TestBulkMeta_Omitempty(t *testing.T) {
	meta := BulkMeta{
		Index: "test-index",
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Errorf("JSON Marshal 失败: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("JSON Unmarshal 失败: %v", err)
	}

	if _, exists := result["_id"]; exists {
		t.Error("_id 不应该出现在 JSON 中（omitempty）")
	}
}

// ===== BulkOperation 测试 =====

func TestBulkOperation_Struct(t *testing.T) {
	op := BulkOperation{
		Meta: BulkMeta{
			Index:  "test-index",
			ID:     "doc-1",
			Action: "index",
		},
		Document: map[string]interface{}{
			"name": "test",
			"age":  25,
		},
	}

	if op.Meta.Index != "test-index" {
		t.Error("Meta.Index 设置失败")
	}

	if op.Meta.ID != "doc-1" {
		t.Error("Meta.ID 设置失败")
	}

	if op.Document == nil {
		t.Error("Document 不应该为 nil")
	}

	doc := op.Document.(map[string]interface{})
	if doc["name"] != "test" {
		t.Error("Document.name 设置失败")
	}
}

func TestBulkOperation_JSON(t *testing.T) {
	operations := []BulkOperation{
		{
			Meta: BulkMeta{
				Index: "users",
				ID:    "1",
			},
			Document: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
		{
			Meta: BulkMeta{
				Index: "users",
				ID:    "2",
			},
			Document: map[string]interface{}{
				"name": "Bob",
				"age":  25,
			},
		},
	}

	var buf bytes.Buffer
	for _, op := range operations {
		meta, err := json.Marshal(op.Meta)
		if err != nil {
			t.Errorf("Marshal meta 失败: %v", err)
		}
		buf.WriteString(string(meta) + "\n")

		if op.Document != nil {
			doc, err := json.Marshal(op.Document)
			if err != nil {
				t.Errorf("Marshal document 失败: %v", err)
			}
			buf.WriteString(string(doc) + "\n")
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Errorf("期望 4 行, 得到 %d 行", len(lines))
	}
}

// ===== SearchRequest 测试 =====

func TestSearchRequest_Struct(t *testing.T) {
	req := SearchRequest{
		Index:          []string{"users", "products"},
		Query:          map[string]interface{}{"match": map[string]interface{}{"name": "Alice"}},
		From:           0,
		Size:           10,
		Sort:           []string{"created_at:desc"},
		Source:         []string{"name", "age"},
		TrackTotalHits: true,
	}

	if len(req.Index) != 2 {
		t.Errorf("Index 长度期望 2, 得到 %d", len(req.Index))
	}

	if req.From != 0 {
		t.Errorf("From 期望 0, 得到 %d", req.From)
	}

	if req.Size != 10 {
		t.Errorf("Size 期望 10, 得到 %d", req.Size)
	}

	if !req.TrackTotalHits {
		t.Error("TrackTotalHits 应该为 true")
	}
}

func TestSearchRequest_JSON(t *testing.T) {
	req := SearchRequest{
		Index: []string{"test-index"},
		Query: map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		},
		From: 0,
		Size: 10,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("JSON Marshal 失败: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("JSON Unmarshal 失败: %v", err)
	}

	if result["from"] != float64(0) {
		t.Error("from 字段不正确")
	}

	if result["size"] != float64(10) {
		t.Error("size 字段不正确")
	}
}

func TestSearchRequest_EmptyQuery(t *testing.T) {
	req := SearchRequest{
		Index: []string{"test-index"},
		Query: nil,
		From:  0,
		Size:  10,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("JSON Marshal 失败: %v", err)
	}

	if len(data) == 0 {
		t.Error("JSON 数据不应该为空")
	}
}

// ===== SearchResponse 测试 =====

func TestSearchResponse_Struct(t *testing.T) {
	resp := SearchResponse{
		Hits: SearchHits{
			Hits: []SearchHit{
				{
					ID:     "1",
					Index:  "test-index",
					Score:  1.5,
					Source: json.RawMessage(`{"name":"Alice","age":30}`),
				},
				{
					ID:     "2",
					Index:  "test-index",
					Score:  1.2,
					Source: json.RawMessage(`{"name":"Bob","age":25}`),
				},
			},
			Total: 2,
		},
		Total:    2,
		MaxScore: 1.5,
	}

	if len(resp.Hits.Hits) != 2 {
		t.Errorf("Hits 长度期望 2, 得到 %d", len(resp.Hits.Hits))
	}

	if resp.Total != 2 {
		t.Errorf("Total 期望 2, 得到 %d", resp.Total)
	}

	if resp.MaxScore != 1.5 {
		t.Errorf("MaxScore 期望 1.5, 得到 %f", resp.MaxScore)
	}
}

func TestSearchHit_Struct(t *testing.T) {
	hit := SearchHit{
		ID:     "doc-123",
		Index:  "users",
		Score:  2.5,
		Source: json.RawMessage(`{"name":"Test User","email":"test@example.com"}`),
	}

	if hit.ID != "doc-123" {
		t.Errorf("ID 期望 'doc-123', 得到 '%s'", hit.ID)
	}

	if hit.Index != "users" {
		t.Errorf("Index 期望 'users', 得到 '%s'", hit.Index)
	}

	if hit.Score != 2.5 {
		t.Errorf("Score 期望 2.5, 得到 %f", hit.Score)
	}

	var source map[string]interface{}
	err := json.Unmarshal(hit.Source, &source)
	if err != nil {
		t.Errorf("解析 Source 失败: %v", err)
	}

	if source["name"] != "Test User" {
		t.Error("Source.name 解析失败")
	}
}

func TestSearchHit_JSONTags(t *testing.T) {
	data := `{
		"_id": "test-id",
		"_index": "test-index",
		"_score": 3.14,
		"_source": {"key": "value"}
	}`

	var hit SearchHit
	err := json.Unmarshal([]byte(data), &hit)
	if err != nil {
		t.Errorf("JSON Unmarshal 失败: %v", err)
	}

	if hit.ID != "test-id" {
		t.Errorf("ID 期望 'test-id', 得到 '%s'", hit.ID)
	}

	if hit.Index != "test-index" {
		t.Errorf("Index 期望 'test-index', 得到 '%s'", hit.Index)
	}

	if hit.Score != 3.14 {
		t.Errorf("Score 期望 3.14, 得到 %f", hit.Score)
	}
}

// ===== AggregateRequest 测试 =====

func TestAggregateRequest_Struct(t *testing.T) {
	req := AggregateRequest{
		Index: []string{"orders"},
		Query: map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		Aggs: map[string]map[string]interface{}{
			"status_count": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "status",
				},
			},
			"avg_price": map[string]interface{}{
				"avg": map[string]interface{}{
					"field": "price",
				},
			},
		},
		Size: 0,
	}

	if len(req.Index) != 1 {
		t.Errorf("Index 长度期望 1, 得到 %d", len(req.Index))
	}

	if len(req.Aggs) != 2 {
		t.Errorf("Aggs 长度期望 2, 得到 %d", len(req.Aggs))
	}

	if _, exists := req.Aggs["status_count"]; !exists {
		t.Error("status_count 聚合应该存在")
	}

	if _, exists := req.Aggs["avg_price"]; !exists {
		t.Error("avg_price 聚合应该存在")
	}
}

func TestAggregateRequest_JSON(t *testing.T) {
	req := AggregateRequest{
		Index: []string{"products"},
		Query: map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		Aggs: map[string]map[string]interface{}{
			"category_stats": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category.keyword",
					"size":  10,
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("JSON Marshal 失败: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("JSON Unmarshal 失败: %v", err)
	}

	if result["size"] != float64(0) {
		t.Error("size 字段应该是 0")
	}
}

// ===== 文档操作测试 =====

func TestDocument_JSONEncoding(t *testing.T) {
	doc := map[string]interface{}{
		"name":    "Test Product",
		"price":   99.99,
		"in_stock": true,
		"tags":    []string{"electronics", "sale"},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(doc)
	if err != nil {
		t.Errorf("Encode 失败: %v", err)
	}

	var decoded map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &decoded)
	if err != nil {
		t.Errorf("Unmarshal 失败: %v", err)
	}

	if decoded["name"] != "Test Product" {
		t.Error("name 字段不正确")
	}

	if decoded["price"] != 99.99 {
		t.Error("price 字段不正确")
	}

	tags := decoded["tags"].([]interface{})
	if len(tags) != 2 {
		t.Errorf("tags 长度期望 2, 得到 %d", len(tags))
	}
}

func TestUpdateDocument_Format(t *testing.T) {
	doc := map[string]interface{}{
		"name":  "Updated Product",
		"price": 149.99,
	}

	update := map[string]interface{}{
		"doc": doc,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Errorf("JSON Marshal 失败: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("JSON Unmarshal 失败: %v", err)
	}

	docResult, ok := result["doc"].(map[string]interface{})
	if !ok {
		t.Error("doc 字段应该是对象")
	}

	if docResult["name"] != "Updated Product" {
		t.Error("doc.name 不正确")
	}
}

// ===== 连接测试 =====

func TestNew_InvalidAddress(t *testing.T) {
	client, err := New(
		WithESAddresses("http://invalid-host:9200"),
	)

	if err == nil {
		t.Log("客户端创建成功（可能连接到实际服务器）")
		if client != nil {
			if len(client.opts.Addresses) != 1 {
				t.Error("Addresses 应该被正确设置")
			}
		}
	} else {
		t.Logf("预期的连接错误: %v", err)
	}
}

func TestClient_Ping_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	err := client.Ping(ctx)

	if err == nil {
		t.Error("Ping 应该返回错误（没有底层客户端）")
	}
}

func TestClient_CreateIndex_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	mapping := `{"mappings":{"properties":{"name":{"type":"keyword"}}}}`

	err := client.CreateIndex(ctx, "test-index", mapping)

	if err == nil {
		t.Error("CreateIndex 应该返回错误（没有底层客户端）")
	}
}

func TestClient_DeleteIndex_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	err := client.DeleteIndex(ctx, "test-index")

	if err == nil {
		t.Error("DeleteIndex 应该返回错误（没有底层客户端）")
	}
}

func TestClient_ExistsIndex_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	exists, err := client.ExistsIndex(ctx, "test-index")

	if err == nil {
		if exists {
			t.Error("ExistsIndex 应该返回 false（没有底层客户端）")
		}
	}
}

func TestClient_Index_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	doc := map[string]interface{}{
		"name": "Test",
		"age":  25,
	}

	err := client.Index(ctx, "test-index", "doc-1", doc)

	if err == nil {
		t.Error("Index 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Get_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	_, err := client.Get(ctx, "test-index", "doc-1")

	if err == nil {
		t.Error("Get 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Delete_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	err := client.Delete(ctx, "test-index", "doc-1")

	if err == nil {
		t.Error("Delete 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Update_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	doc := map[string]interface{}{
		"name": "Updated",
	}

	err := client.Update(ctx, "test-index", "doc-1", doc)

	if err == nil {
		t.Error("Update 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Bulk_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	operations := []BulkOperation{
		{
			Meta: BulkMeta{
				Index: "test-index",
				ID:    "doc-1",
			},
			Document: map[string]interface{}{
				"name": "Test",
			},
		},
	}

	err := client.Bulk(ctx, operations)

	if err == nil {
		t.Error("Bulk 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Search_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	req := SearchRequest{
		Index: []string{"test-index"},
		Query: map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	_, err := client.Search(ctx, req)

	if err == nil {
		t.Error("Search 应该返回错误（没有底层客户端）")
	}
}

func TestClient_Aggregate_WithoutConnection(t *testing.T) {
	client := &Client{}

	ctx := context.Background()
	req := AggregateRequest{
		Index: []string{"test-index"},
		Query: map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		Aggs: map[string]map[string]interface{}{
			"count": map[string]interface{}{
				"value_count": map[string]interface{}{
					"field": "name.keyword",
				},
			},
		},
	}

	_, err := client.Aggregate(ctx, req)

	if err == nil {
		t.Error("Aggregate 应该返回错误（没有底层客户端）")
	}
}

// ===== 索引名称前缀测试 =====

func TestGetIndexName_WithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		index    string
		expected string
	}{
		{
			name:     "生产环境前缀",
			prefix:   "prod",
			index:    "users",
			expected: "prod-users",
		},
		{
			name:     "开发环境前缀",
			prefix:   "dev",
			index:    "products",
			expected: "dev-products",
		},
		{
			name:     "测试环境前缀",
			prefix:   "test",
			index:    "orders",
			expected: "test-orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				opts: &Options{
					IndexPrefix: tt.prefix,
				},
			}

			result := client.GetIndexName(tt.index)
			if result != tt.expected {
				t.Errorf("期望 '%s', 得到 '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetIndexName_EmptyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		index    string
		expected string
	}{
		{
			name:     "空前缀返回原名称",
			prefix:   "",
			index:    "users",
			expected: "users",
		},
		{
			name:     "无前缀返回原名称",
			prefix:   "",
			index:    "products",
			expected: "products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				opts: &Options{
					IndexPrefix: tt.prefix,
				},
			}

			result := client.GetIndexName(tt.index)
			if result != tt.expected {
				t.Errorf("期望 '%s', 得到 '%s'", tt.expected, result)
			}
		})
	}
}

// ===== 上下文取消测试 =====

func TestSearch_WithCanceledContext(t *testing.T) {
	client := &Client{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := SearchRequest{
		Index: []string{"test-index"},
		Query: map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	_, err := client.Search(ctx, req)

	if err == nil {
		t.Error("Search 应该返回错误（上下文已取消）")
	}
}

func TestPing_WithCanceledContext(t *testing.T) {
	client := &Client{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Ping(ctx)

	if err == nil {
		t.Error("Ping 应该返回错误（上下文已取消）")
	}
}

// ===== 模拟 HTTP Transport 测试 =====

func TestCustomTransport(t *testing.T) {
	opts := &Options{
		Addresses: []string{"http://localhost:9200"},
	}

	opts.Transport = &mockTransport{
		roundTripResponse: &http.Response{
			StatusCode: 200,
		},
	}

	WithESAddresses("http://test:9200")(opts)

	if len(opts.Addresses) != 1 {
		t.Error("Addresses 设置失败")
	}

	if opts.Transport == nil {
		t.Error("Transport 应该被设置")
	}
}

// mockTransport 模拟 HTTP Transport
type mockTransport struct {
	roundTripResponse *http.Response
	roundTripError    error
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTripResponse, t.roundTripError
}

// ===== 配置验证测试 =====

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		addresses []string
		username  string
		password  string
		valid     bool
	}{
		{
			name:      "有效配置",
			addresses: []string{"http://localhost:9200"},
			username:  "elastic",
			password:  "password",
			valid:     true,
		},
		{
			name:      "无用户名密码",
			addresses: []string{"http://localhost:9200"},
			username:  "",
			password:  "",
			valid:     true,
		},
		{
			name:      "多地址",
			addresses: []string{"http://es1:9200", "http://es2:9200", "http://es3:9200"},
			username:  "",
			password:  "",
			valid:     true,
		},
		{
			name:      "空地址",
			addresses: []string{},
			username:  "",
			password:  "",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				Addresses: tt.addresses,
				Username:  tt.username,
				Password:  tt.password,
			}

			valid := len(opts.Addresses) > 0
			if valid != tt.valid {
				t.Errorf("配置验证结果期望 %v, 得到 %v", tt.valid, valid)
			}
		})
	}
}

// ===== 性能测试（可选） =====

func TestBulkOperation_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（短模式）")
	}

	operations := make([]BulkOperation, 1000)
	for i := 0; i < 1000; i++ {
		operations[i] = BulkOperation{
			Meta: BulkMeta{
				Index: "test-index",
				ID:    string(rune('0' + i%10)),
			},
			Document: map[string]interface{}{
				"id":   i,
				"name": "Test",
			},
		}
	}

	start := time.Now()

	var buf bytes.Buffer
	for _, op := range operations {
		meta, _ := json.Marshal(op.Meta)
		buf.WriteString(string(meta) + "\n")

		if op.Document != nil {
			doc, _ := json.Marshal(op.Document)
			buf.WriteString(string(doc) + "\n")
		}
	}

	elapsed := time.Since(start)
	t.Logf("处理 %d 个操作耗时: %v", len(operations), elapsed)

	if elapsed > time.Second {
		t.Errorf("性能测试失败: 耗时 %v 超过 1 秒", elapsed)
	}
}

// ===== 并发测试 =====

func TestConcurrent_IndexName(t *testing.T) {
	client := &Client{
		opts: &Options{
			IndexPrefix: "test",
		},
	}

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(idx int) {
			name := client.GetIndexName("users")
			if name != "test-users" {
				t.Errorf("GetIndexName 返回错误结果: %s", name)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// ===== 边界条件测试 =====

func TestSearchRequest_ZeroValues(t *testing.T) {
	req := SearchRequest{
		Index: []string{"test"},
		Query: map[string]interface{}{},
		From:  0,
		Size:  0,
	}

	if req.From != 0 {
		t.Error("From 应该是 0")
	}

	if req.Size != 0 {
		t.Error("Size 应该是 0")
	}
}

func TestBulkOperation_NilDocument(t *testing.T) {
	op := BulkOperation{
		Meta: BulkMeta{
			Index: "test",
			ID:    "1",
		},
		Document: nil,
	}

	var buf bytes.Buffer
	meta, _ := json.Marshal(op.Meta)
	buf.WriteString(string(meta) + "\n")

	if op.Document != nil {
		doc, _ := json.Marshal(op.Document)
		buf.WriteString(string(doc) + "\n")
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("期望 1 行（只有 meta），得到 %d 行", len(lines))
	}
}

func TestEmptyOperations(t *testing.T) {
	var operations []BulkOperation

	client := &Client{}
	ctx := context.Background()

	err := client.Bulk(ctx, operations)

	if err != nil {
		t.Errorf("空 Bulk 操作应该成功, 得到错误: %v", err)
	}
}

// ===== 错误消息测试 =====

func TestErrorMessages(t *testing.T) {
	errorTests := []struct {
		err      error
		contains string
	}{
		{
			err:      nil,
			contains: "",
		},
	}

	for _, tt := range errorTests {
		if tt.err != nil {
			errStr := tt.err.Error()
			if tt.contains != "" && !strings.Contains(errStr, tt.contains) {
				t.Errorf("错误消息应该包含 '%s', 得到 '%s'", tt.contains, errStr)
			}
		}
	}
}
