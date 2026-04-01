package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/pkg/log"
)

// GinServer Gin HTTP 服务器实现
// 基于 Gin 框架的 HTTP 服务器，提供高性能的路由和中间件支持
type GinServer struct {
	name         string
	engine       *gin.Engine
	httpServer   *http.Server
	opts         *Options
	middlewares  []gin.HandlerFunc
	RouterGroups []gin.RouterGroup
}

// NewGinServer 创建新的 Gin 服务器实例
// name: 服务器名称
// opts: 可选的配置选项
func NewGinServer(name string, opts ...Option) *GinServer {
	options := NewOptions()

	for _, opt := range opts {
		opt(options)
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// 添加全局中间件
	globalMiddlewares := []gin.HandlerFunc{
		GinRecovery(),
		GinLogger(),
	}
	engine.Use(globalMiddlewares...)

	// 添加自定义中间件
	for _, m := range options.Middleware {
		engine.Use(ToGinMiddleware(m))
	}

	server := &GinServer{
		name:   name,
		engine: engine,
		opts:   options,
	}

	// 创建 HTTP 服务器
	server.httpServer = &http.Server{
		Addr:         options.Addr,
		Handler:      engine,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		IdleTimeout:  options.IdleTimeout,
	}

	return server
}

// Name 获取服务器名称
func (s *GinServer) Name() string {
	return s.name
}

// Start 启动 Gin 服务器
func (s *GinServer) Start() error {
	ln, err := net.Listen(s.opts.Network, s.opts.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Info("gin server starting",
		log.String("name", s.name),
		log.String("addr", s.opts.Addr))

	return s.httpServer.Serve(ln)
}

// Stop 停止 Gin 服务器
func (s *GinServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("gin server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// Engine 获取 Gin 引擎实例
// 返回 *gin.Engine，可用于直接访问 Gin 的高级功能
func (s *GinServer) Engine() *gin.Engine {
	return s.engine
}

// HttpServer 获取底层 HTTP 服务器
// 返回 *http.Server，可用于获取服务器配置信息
func (s *GinServer) HttpServer() *http.Server {
	return s.httpServer
}

// GET 注册 GET 路由
func (s *GinServer) GET(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.GET(relativePath, ginHandlers...)
}

// POST 注册 POST 路由
func (s *GinServer) POST(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.POST(relativePath, ginHandlers...)
}

// PUT 注册 PUT 路由
func (s *GinServer) PUT(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.PUT(relativePath, ginHandlers...)
}

// DELETE 注册 DELETE 路由
func (s *GinServer) DELETE(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.DELETE(relativePath, ginHandlers...)
}

// PATCH 注册 PATCH 路由
func (s *GinServer) PATCH(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.PATCH(relativePath, ginHandlers...)
}

// OPTIONS 注册 OPTIONS 路由
func (s *GinServer) OPTIONS(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.OPTIONS(relativePath, ginHandlers...)
}

// HEAD 注册 HEAD 路由
func (s *GinServer) HEAD(relativePath string, handlers ...interface{}) {
	ginHandlers := s.convertHandlers(handlers)
	s.engine.HEAD(relativePath, ginHandlers...)
}

// Group 创建路由组
func (s *GinServer) Group(relativePath string, handlers ...interface{}) interface{} {
	if len(handlers) > 0 {
		ginHandlers := s.convertHandlers(handlers)
		return s.engine.Group(relativePath, ginHandlers...)
	}
	return s.engine.Group(relativePath)
}

// Use 注册中间件
func (s *GinServer) Use(middleware ...interface{}) {
	ginMiddlewares := s.convertHandlers(middleware)
	s.engine.Use(ginMiddlewares...)
}

// convertHandlers 将通用接口转换为 Gin 处理器
func (s *GinServer) convertHandlers(handlers []interface{}) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, 0, len(handlers))
	for _, h := range handlers {
		if hf, ok := h.(gin.HandlerFunc); ok {
			result = append(result, hf)
		}
	}
	return result
}

// GinRecovery Gin 恢复中间件
// 捕获 panic 并返回 500 错误
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					log.Any("error", err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// GinLogger Gin 日志中间件
// 记录请求方法和路径
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("request",
			log.String("method", method),
			log.String("path", path),
			log.Int("status", status),
			log.Duration("latency", latency))
	}
}

// ToGinMiddleware 将标准中间件转换为 Gin 中间件
func ToGinMiddleware(m Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

// GinTransportFactory Gin 传输工厂
type GinTransportFactory struct{}

// NewGinTransportFactory 创建 Gin 传输工厂
func NewGinTransportFactory() *GinTransportFactory {
	return &GinTransportFactory{}
}

// CreateServer 创建 Gin 服务器
func (f *GinTransportFactory) CreateServer(name string, opts ...Option) (Server, error) {
	return NewGinServer(name, opts...), nil
}

// GetType 获取传输类型
func (f *GinTransportFactory) GetType() TransportType {
	return TransportTypeGin
}
