package server

import (
	"context"
	"testing"
	"time"

	"github.com/jiujuan/go-ants/pkg/transport"
	"github.com/stretchr/testify/assert"
)

// ===== HTTPServer 测试 =====

func TestNewHTTPServer(t *testing.T) {
	// 创建 mock transport.Server
	mockServer := &mockTransportServer{
		name: "test-server",
	}

	server := NewHTTPServer("test-http", mockServer)

	assert.NotNil(t, server)
	assert.Equal(t, "test-http", server.Name())
	assert.Equal(t, mockServer, server.server)
}

func TestHTTPServerName(t *testing.T) {
	mockServer := &mockTransportServer{name: "mock"}
	server := NewHTTPServer("my-server", mockServer)

	assert.Equal(t, "my-server", server.Name())
}

func TestHTTPServerStart(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: false,
	}

	server := NewHTTPServer("start-test", mockServer)
	ctx := context.Background()

	err := server.Start(ctx)

	assert.NoError(t, err)
	assert.True(t, mockServer.started)
	assert.True(t, mockServer.stopped == false)
}

func TestHTTPServerStop(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: false,
	}

	server := NewHTTPServer("stop-test", mockServer)
	ctx := context.Background()

	err := server.Stop(ctx)

	assert.NoError(t, err)
	assert.True(t, mockServer.stopped)
}

func TestHTTPServerEngine(t *testing.T) {
	mockServer := &mockTransportServer{name: "test"}
	server := NewHTTPServer("engine-test", mockServer)

	// 初始状态
	assert.Nil(t, server.Engine())

	// 设置引擎
	engine := "mock-engine"
	server.SetEngine(engine)

	assert.Equal(t, engine, server.Engine())
}

func TestHTTPServerSetEngine(t *testing.T) {
	mockServer := &mockTransportServer{name: "test"}
	server := NewHTTPServer("set-engine-test", mockServer)

	// 测试设置不同类型的引擎
	ginEngine := make(map[string]interface{})
	server.SetEngine(ginEngine)
	assert.Equal(t, ginEngine, server.Engine())

	fiberEngine := "fiber-engine"
	server.SetEngine(fiberEngine)
	assert.Equal(t, fiberEngine, server.Engine())
}

// ===== HTTPServerBuilder 测试 =====

func TestNewHTTPServerBuilder(t *testing.T) {
	builder := NewHTTPServerBuilder()

	assert.NotNil(t, builder)
	assert.Equal(t, "http", builder.name)
	assert.Equal(t, ":8080", builder.address)
	assert.NotNil(t, builder.middleware)
	assert.Empty(t, builder.middleware)
}

func TestHTTPServerBuilderWithName(t *testing.T) {
	builder := NewHTTPServerBuilder()

	result := builder.WithName("custom-server")

	assert.Equal(t, builder, result) // 验证链式调用
	assert.Equal(t, "custom-server", builder.name)
}

func TestHTTPServerBuilderWithAddress(t *testing.T) {
	builder := NewHTTPServerBuilder()

	result := builder.WithAddress("localhost:9090")

	assert.Equal(t, builder, result)
	assert.Equal(t, "localhost:9090", builder.address)
}

func TestHTTPServerBuilderWithAddressDifferentFormats(t *testing.T) {
	testCases := []struct {
		name    string
		addr    string
	}{
		{"default", ":8080"},
		{"localhost", "localhost:8080"},
		{"ip", "192.168.1.1:8080"},
		{"custom port", ":3000"},
		{"unix socket", "/tmp/server.sock"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewHTTPServerBuilder()
			builder.WithAddress(tc.addr)
			assert.Equal(t, tc.addr, builder.address)
		})
	}
}

func TestHTTPServerBuilderWithMiddleware(t *testing.T) {
	builder := NewHTTPServerBuilder()

	// 添加第一个中间件
	result := builder.WithMiddleware("mw1")
	assert.Equal(t, builder, result)
	assert.Len(t, builder.middleware, 1)
	assert.Equal(t, "mw1", builder.middleware[0])

	// 添加多个中间件
	builder.WithMiddleware("mw2", "mw3")
	assert.Len(t, builder.middleware, 3)
}

func TestHTTPServerBuilderWithMiddlewareChain(t *testing.T) {
	builder := NewHTTPServerBuilder()

	builder.
		WithName("api-server").
		WithAddress(":8888").
		WithMiddleware("cors").
		WithMiddleware("logger", "auth")

	assert.Equal(t, "api-server", builder.name)
	assert.Equal(t, ":8888", builder.address)
	assert.Len(t, builder.middleware, 3)
}

func TestHTTPServerBuilderBuildGin(t *testing.T) {
	builder := NewHTTPServerBuilder()
	builder.WithName("gin-server").WithAddress(":8081")

	server := builder.BuildGin()

	assert.NotNil(t, server)
	assert.Equal(t, "gin-server", server.Name())
}

func TestHTTPServerBuilderBuildFiber(t *testing.T) {
	builder := NewHTTPServerBuilder()
	builder.WithName("fiber-server").WithAddress(":8082")

	server := builder.BuildFiber()

	assert.NotNil(t, server)
	assert.Equal(t, "fiber-server", server.Name())
}

// ===== Server 测试 =====

func TestNewServer(t *testing.T) {
	mockServer := &mockTransportServer{name: "mock"}
	httpServer := NewHTTPServer("test", mockServer)

	server := New(httpServer)

	assert.NotNil(t, server)
	assert.Equal(t, httpServer, server.httpserver)
}

func TestServerName(t *testing.T) {
	mockServer := &mockTransportServer{name: "mock"}
	httpServer := NewHTTPServer("test", mockServer)
	server := New(httpServer)

	assert.Equal(t, "server", server.Name())
}

func TestServerStart(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: false,
	}
	httpServer := NewHTTPServer("test", mockServer)
	server := New(httpServer)

	ctx := context.Background()
	err := server.Start(ctx)

	assert.NoError(t, err)
	assert.True(t, mockServer.started)
}

func TestServerStop(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: false,
	}
	httpServer := NewHTTPServer("test", mockServer)
	server := New(httpServer)

	ctx := context.Background()
	err := server.Stop(ctx)

	assert.NoError(t, err)
	assert.True(t, mockServer.stopped)
}

// ===== Wire Provider 测试 =====

func TestNewGinServer(t *testing.T) {
	server := NewGinServer(":8085")

	assert.NotNil(t, server)
	assert.Equal(t, "gin", server.Name())
}

func TestNewFiberServer(t *testing.T) {
	server := NewFiberServer(":8086")

	assert.NotNil(t, server)
	assert.Equal(t, "fiber", server.Name())
}

func TestProviderSet(t *testing.T) {
	// 验证 ProviderSet 包含预期的提供者
	assert.NotNil(t, ProviderSet)
}

// ===== 集成场景测试 =====

func TestFullServerLifecycle(t *testing.T) {
	// 使用构建器创建服务器
	builder := NewHTTPServerBuilder()
	builder.
		WithName("lifecycle-test").
		WithAddress(":0"). // 使用端口 0 让系统分配端口
		WithMiddleware("cors")

	httpServer := builder.BuildGin()
	server := New(httpServer)

	ctx := context.Background()

	// 启动服务器
	err := server.Start(ctx)
	assert.NoError(t, err)

	// 停止服务器
	err = server.Stop(ctx)
	assert.NoError(t, err)
}

func TestGinAndFiberServers(t *testing.T) {
	// 测试 Gin 服务器
	ginServer := NewHTTPServerBuilder().
		WithName("gin-api").
		WithAddress(":8083").
		BuildGin()

	assert.Equal(t, "gin-api", ginServer.Name())

	// 测试 Fiber 服务器
	fiberServer := NewHTTPServerBuilder().
		WithName("fiber-api").
		WithAddress(":8084").
		BuildFiber()

	assert.Equal(t, "fiber-api", fiberServer.Name())
}

// ===== 错误处理测试 =====

func TestHTTPServerStartError(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: true,
		startErr:   assert.AnError,
	}

	server := NewHTTPServer("error-test", mockServer)
	ctx := context.Background()

	err := server.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestHTTPServerStopError(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "test",
		shouldFail: true,
		stopErr:    assert.AnError,
	}

	server := NewHTTPServer("stop-error-test", mockServer)
	ctx := context.Background()

	err := server.Stop(ctx)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// ===== 并发测试 =====

func TestHTTPServerConcurrentAccess(t *testing.T) {
	mockServer := &mockTransportServer{name: "test"}
	server := NewHTTPServer("concurrent-test", mockServer)

	// 并发访问 Name() 方法
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			_ = server.Name()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestHTTPServerConcurrentEngineAccess(t *testing.T) {
	mockServer := &mockTransportServer{name: "test"}
	server := NewHTTPServer("engine-concurrent-test", mockServer)

	done := make(chan bool)

	// 并发设置引擎
	for i := 0; i < 50; i++ {
		go func(id int) {
			server.SetEngine("engine-" + string(rune('0'+id%10)))
			done <- true
		}(i)
	}

	// 并发获取引擎
	for i := 0; i < 50; i++ {
		go func() {
			_ = server.Engine()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// ===== 边界条件测试 =====

func TestHTTPServerBuilderEmptyName(t *testing.T) {
	builder := NewHTTPServerBuilder()
	builder.WithName("")

	// 空名称应该被接受
	assert.Equal(t, "", builder.name)
}

func TestHTTPServerBuilderEmptyAddress(t *testing.T) {
	builder := NewHTTPServerBuilder()
	builder.WithAddress("")

	assert.Equal(t, "", builder.address)
}

func TestHTTPServerBuilderManyMiddleware(t *testing.T) {
	builder := NewHTTPServerBuilder()

	// 添加大量中间件
	middlewares := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		middlewares[i] = "mw" + string(rune('0'+i%10))
	}

	builder.WithMiddleware(middlewares...)

	assert.Len(t, builder.middleware, 100)
}

// ===== mock 实现 =====

// mockTransportServer 模拟 transport.Server 实现
type mockTransportServer struct {
	name       string
	started    bool
	stopped    bool
	shouldFail bool
	startErr   error
	stopErr    error
}

func (m *mockTransportServer) Start() error {
	m.started = true
	if m.shouldFail && m.startErr != nil {
		return m.startErr
	}
	return nil
}

func (m *mockTransportServer) Stop() error {
	m.stopped = true
	if m.shouldFail && m.stopErr != nil {
		return m.stopErr
	}
	return nil
}

func (m *mockTransportServer) Name() string {
	return m.name
}

// 确保 mock 实现 transport.Server 接口
var _ transport.Server = (*mockTransportServer)(nil)

// ===== 性能测试 =====

func BenchmarkHTTPServerName(b *testing.B) {
	mockServer := &mockTransportServer{name: "benchmark"}
	server := NewHTTPServer("benchmark-test", mockServer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.Name()
	}
}

func BenchmarkHTTPServerStart(b *testing.B) {
	mockServer := &mockTransportServer{name: "benchmark"}
	server := NewHTTPServer("benchmark-test", mockServer)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockServer.started = false
		_ = server.Start(ctx)
	}
}

func BenchmarkHTTPServerStop(b *testing.B) {
	mockServer := &mockTransportServer{name: "benchmark"}
	server := NewHTTPServer("benchmark-test", mockServer)
	ctx := context.Background()

	// 先启动
	_ = server.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockServer.stopped = false
		_ = server.Stop(ctx)
	}
}

// ===== 辅助测试函数 =====

func TestMockTransportServerInterface(t *testing.T) {
	// 确保 mock 实现了 transport.Server 接口
	var _ transport.Server = (*mockTransportServer)(nil)

	server := &mockTransportServer{name: "interface-test"}
	assert.Equal(t, "interface-test", server.Name())
}

func TestHTTPServerInterface(t *testing.T) {
	// 确保 HTTPServer 实现了 app.Component 接口
	// 注意：这需要在 app 包存在的情况下运行
	// var _ app.Component = (*HTTPServer)(nil)
	// var _ app.NamedComponent = (*HTTPServer)(nil)

	// 这里我们只验证基本功能
	mockServer := &mockTransportServer{name: "test"}
	httpServer := NewHTTPServer("interface-test", mockServer)

	assert.NotNil(t, httpServer)
	assert.Equal(t, "interface-test", httpServer.Name())
}

// 测试超时场景
func TestHTTPServerContextTimeout(t *testing.T) {
	mockServer := &mockTransportServer{
		name:       "timeout-test",
		shouldFail: false,
	}
	server := NewHTTPServer("timeout-server", mockServer)

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// 启动服务器（虽然不会真正启动，但测试 context）
	err := server.Start(ctx)
	assert.NoError(t, err)
}
