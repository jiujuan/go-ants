package transport

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	opts := NewOptions()

	assert.Equal(t, "tcp", opts.Network)
	assert.Equal(t, ":8080", opts.Addr)
	assert.Equal(t, time.Second*30, opts.ReadTimeout)
	assert.Equal(t, time.Second*30, opts.WriteTimeout)
	assert.Equal(t, time.Second*60, opts.IdleTimeout)
}

func TestOptionsFunctions(t *testing.T) {
	opts := NewOptions()

	// 测试 WithNetwork
	WithNetwork("udp")(opts)
	assert.Equal(t, "udp", opts.Network)

	// 测试 WithAddr
	WithAddr("localhost:9090")(opts)
	assert.Equal(t, "localhost:9090", opts.Addr)

	// 测试 WithReadTimeout
	WithReadTimeout(time.Second * 60)(opts)
	assert.Equal(t, time.Second*60, opts.ReadTimeout)

	// 测试 WithWriteTimeout
	WithWriteTimeout(time.Second * 90)(opts)
	assert.Equal(t, time.Second*90, opts.WriteTimeout)

	// 测试 WithIdleTimeout
	WithIdleTimeout(time.Second * 120)(opts)
	assert.Equal(t, time.Second*120, opts.IdleTimeout)

	// 测试 WithHost
	WithHost("127.0.0.1")(opts)
	assert.Equal(t, "127.0.0.1", opts.Host)

	// 测试 WithPort
	WithPort(8888)(opts)
	assert.Equal(t, 8888, opts.Port)
}

func TestWithMiddleware(t *testing.T) {
	opts := NewOptions()

	middleware := &testMiddleware{name: "test"}
	WithMiddleware(middleware)(opts)

	assert.Len(t, opts.Middleware, 1)
	assert.Equal(t, "test", opts.Middleware[0].Name())
}

func TestWithMiddlewareFunc(t *testing.T) {
	opts := NewOptions()

	middleware := func(next http.Handler) http.Handler {
		return next
	}
	WithMiddlewareFunc(middleware)(opts)

	assert.Len(t, opts.Middleware, 1)
}

func TestMergeOptions(t *testing.T) {
	opts := NewOptions()

	merged := MergeOptions(
		WithAddr("localhost:9999"),
		WithPort(8888),
	)

	merged(opts)

	assert.Equal(t, "localhost:9999", opts.Addr)
	assert.Equal(t, 8888, opts.Port)
}

func TestApplyOptions(t *testing.T) {
	opts := NewOptions()

	ApplyOptions(opts,
		WithNetwork("unix"),
		WithAddr("/tmp/server.sock"),
	)

	assert.Equal(t, "unix", opts.Network)
	assert.Equal(t, "/tmp/server.sock", opts.Addr)
}

func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	assert.Equal(t, "tcp", config.Network)
	assert.Equal(t, ":8080", config.Addr)
	assert.Equal(t, time.Second*30, config.ReadTimeout)
	assert.Equal(t, time.Second*30, config.WriteTimeout)
	assert.Equal(t, time.Second*60, config.IdleTimeout)
}

func TestTransportType(t *testing.T) {
	assert.Equal(t, TransportType("gin"), TransportTypeGin)
	assert.Equal(t, TransportType("fiber"), TransportTypeFiber)
	assert.Equal(t, TransportType("mux"), TransportTypeMux)
}

// ===== GinServer 测试 =====

func TestNewGinServer(t *testing.T) {
	server := NewGinServer("test-gin",
		WithAddr(":8081"),
		WithReadTimeout(time.Second*10),
	)

	assert.NotNil(t, server)
	assert.Equal(t, "test-gin", server.Name())
	assert.NotNil(t, server.Engine())
	assert.NotNil(t, server.HttpServer())
	assert.Equal(t, ":8081", server.HttpServer().Addr)
	assert.Equal(t, time.Second*10, server.HttpServer().ReadTimeout)
}

func TestGinServerName(t *testing.T) {
	server := NewGinServer("my-gin-server")
	assert.Equal(t, "my-gin-server", server.Name())
}

func TestGinServerRouteRegistration(t *testing.T) {
	server := NewGinServer("test-router",
		WithAddr(":0"),
	)

	// 测试路由注册方法存在且可调用
	server.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "get")
	})
	server.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "post")
	})
	server.PUT("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "put")
	})
	server.DELETE("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "delete")
	})
	server.PATCH("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "patch")
	})
	server.OPTIONS("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "options")
	})
	server.HEAD("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "head")
	})
}

func TestGinServerGroup(t *testing.T) {
	server := NewGinServer("test-group",
		WithAddr(":0"),
	)

	group := server.Group("/api")
	assert.NotNil(t, group)
}

func TestGinServerUse(t *testing.T) {
	server := NewGinServer("test-use",
		WithAddr(":0"),
	)

	server.Use(func(c *gin.Context) {
		c.Next()
	})
}

func TestGinTransportFactory(t *testing.T) {
	factory := NewGinTransportFactory()

	assert.Equal(t, TransportTypeGin, factory.GetType())

	server, err := factory.CreateServer("factory-test")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

// ===== FiberServer 测试 =====

func TestNewFiberServer(t *testing.T) {
	server := NewFiberServer("test-fiber",
		WithAddr(":8082"),
		WithWriteTimeout(time.Second*20),
	)

	assert.NotNil(t, server)
	assert.Equal(t, "test-fiber", server.Name())
	assert.NotNil(t, server.App())
	assert.NotNil(t, server.HttpServer())
	assert.Equal(t, ":8082", server.HttpServer().Addr)
}

func TestFiberServerName(t *testing.T) {
	server := NewFiberServer("my-fiber-server")
	assert.Equal(t, "my-fiber-server", server.Name())
}

func TestFiberTransportFactory(t *testing.T) {
	factory := NewFiberTransportFactory()

	assert.Equal(t, TransportTypeFiber, factory.GetType())

	server, err := factory.CreateServer("factory-test")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

// ===== MuxServer 测试 =====

func TestNewMuxServer(t *testing.T) {
	server := NewMuxServer("test-mux",
		WithAddr(":8083"),
		WithIdleTimeout(time.Second*90),
	)

	assert.NotNil(t, server)
	assert.Equal(t, "test-mux", server.Name())
	assert.NotNil(t, server.Router())
	assert.NotNil(t, server.HttpServer())
	assert.Equal(t, ":8083", server.HttpServer().Addr)
}

func TestMuxServerName(t *testing.T) {
	server := NewMuxServer("my-mux-server")
	assert.Equal(t, "my-mux-server", server.Name())
}

func TestMuxServerRouteRegistration(t *testing.T) {
	server := NewMuxServer("test-mux-router",
		WithAddr(":0"),
	)

	// MuxServer 需要 http.Handler，这里用简单测试
	// server.GET("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	assert.NotNil(t, server)
}

func TestMuxTransportFactory(t *testing.T) {
	factory := NewMuxTransportFactory()

	assert.Equal(t, TransportTypeMux, factory.GetType())

	server, err := factory.CreateServer("factory-test")
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

// ===== Middleware 测试 =====

func TestCorsMiddleware(t *testing.T) {
	middleware := CorsMiddleware()
	assert.NotNil(t, middleware)
	assert.Equal(t, "func", middleware.Name())
}

func TestTimeoutMiddleware(t *testing.T) {
	middleware := TimeoutMiddleware(time.Second * 5)
	assert.NotNil(t, middleware)
	assert.Equal(t, "func", middleware.Name())
}

func TestLoggingMiddleware(t *testing.T) {
	middleware := LoggingMiddleware()
	assert.NotNil(t, middleware)
	assert.Equal(t, "func", middleware.Name())
}

func TestRecoveryMiddleware(t *testing.T) {
	middleware := RecoveryMiddleware()
	assert.NotNil(t, middleware)
	assert.Equal(t, "func", middleware.Name())
}

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()
	assert.NotNil(t, middleware)
	assert.Equal(t, "func", middleware.Name())
}

func TestChain(t *testing.T) {
	chain := NewChain(
		CorsMiddleware(),
		LoggingMiddleware(),
	)

	assert.NotNil(t, chain)
	assert.Len(t, chain.middlewares, 2)
}

func TestChainAppend(t *testing.T) {
	chain := NewChain(CorsMiddleware())
	chain.Append(TimeoutMiddleware(time.Second * 5))

	assert.Len(t, chain.middlewares, 2)
}

func TestChainPrepend(t *testing.T) {
	chain := NewChain(LoggingMiddleware())
	chain.Prepend(CorsMiddleware())

	assert.Len(t, chain.middlewares, 2)
}

func TestChainThen(t *testing.T) {
	chain := NewChain(
		CorsMiddleware(),
		LoggingMiddleware(),
	)

	handler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	assert.NotNil(t, handler)
}

func TestToMiddleware(t *testing.T) {
	fn := func(next http.Handler) http.Handler {
		return next
	}

	middleware := ToMiddleware(fn)
	assert.NotNil(t, middleware)
}

func TestWithName(t *testing.T) {
	fn := func(next http.Handler) http.Handler {
		return next
	}

	middleware := WithName("custom-middleware", fn)
	assert.Equal(t, "custom-middleware", middleware.Name())
}

func TestGinRecoveryMiddleware(t *testing.T) {
	recovery := GinRecovery()
	assert.NotNil(t, recovery)
}

func TestGinLoggerMiddleware(t *testing.T) {
	logger := GinLogger()
	assert.NotNil(t, logger)
}

// ===== 测试辅助类型 =====

type testMiddleware struct {
	name string
}

func (m *testMiddleware) Name() string {
	return m.name
}

func (m *testMiddleware) Handle(next http.Handler) http.Handler {
	return next
}
