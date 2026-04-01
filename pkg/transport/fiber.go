package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/jiujuan/go-ants/pkg/log"
)

// FiberServer Fiber HTTP 服务器实现
// 基于 Fiber 框架的高性能 HTTP 服务器
// Fiber 是一个用 Go 编写的 Express-inspired Web 框架，号称是最快的 Go HTTP 框架
type FiberServer struct {
	name        string
	app         *fasthttp.App
	httpServer  *http.Server
	opts        *Options
	middlewares []fasthttp.RequestHandler
}

// NewFiberServer 创建新的 Fiber 服务器实例
// name: 服务器名称
// opts: 可选的配置选项
func NewFiberServer(name string, opts ...Option) *FiberServer {
	options := NewOptions()

	for _, opt := range opts {
		opt(options)
	}

	app := fasthttp.New()

	server := &FiberServer{
		name: name,
		app:  app,
		opts: options,
	}

	// 创建 HTTP 服务器
	server.httpServer = &http.Server{
		Addr:         options.Addr,
		Handler:      app.Handler,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		IdleTimeout:  options.IdleTimeout,
	}

	return server
}

// Name 获取服务器名称
func (s *FiberServer) Name() string {
	return s.name
}

// Start 启动 Fiber 服务器
func (s *FiberServer) Start() error {
	ln, err := net.Listen(s.opts.Network, s.opts.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Info("fiber server starting",
		log.String("name", s.name),
		log.String("addr", s.opts.Addr))

	return s.httpServer.Serve(ln)
}

// Stop 停止 Fiber 服务器
func (s *FiberServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("fiber server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// App 获取 Fiber 应用实例
// 返回 *fasthttp.App，可用于直接访问 Fiber 的高级功能
func (s *FiberServer) App() *fasthttp.App {
	return s.app
}

// HttpServer 获取底层 HTTP 服务器
// 返回 *http.Server，可用于获取服务器配置信息
func (s *FiberServer) HttpServer() *http.Server {
	return s.httpServer
}

// GET 注册 GET 路由
func (s *FiberServer) GET(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.GET(relativePath, fasthttpHandlers...)
}

// POST 注册 POST 路由
func (s *FiberServer) POST(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.POST(relativePath, fasthttpHandlers...)
}

// PUT 注册 PUT 路由
func (s *FiberServer) PUT(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.PUT(relativePath, fasthttpHandlers...)
}

// DELETE 注册 DELETE 路由
func (s *FiberServer) DELETE(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.DELETE(relativePath, fasthttpHandlers...)
}

// PATCH 注册 PATCH 路由
func (s *FiberServer) PATCH(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.PATCH(relativePath, fasthttpHandlers...)
}

// OPTIONS 注册 OPTIONS 路由
func (s *FiberServer) OPTIONS(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.OPTIONS(relativePath, fasthttpHandlers...)
}

// HEAD 注册 HEAD 路由
func (s *FiberServer) HEAD(relativePath string, handlers ...interface{}) {
	fasthttpHandlers := s.convertHandlers(handlers)
	s.app.HEAD(relativePath, fasthttpHandlers...)
}

// Group 创建路由组 (Fiber 中使用 Prefix)
func (s *FiberServer) Group(relativePath string, handlers ...interface{}) interface{} {
	return s.app.Group(relativePath)
}

// Use 注册中间件
func (s *FiberServer) Use(middleware ...interface{}) {
	fasthttpHandlers := s.convertHandlers(middleware)
	s.app.Use(fasthttpHandlers...)
}

// convertHandlers 将通用接口转换为 fasthttp 请求处理器
func (s *FiberServer) convertHandlers(handlers []interface{}) []fasthttp.RequestHandler {
	result := make([]fasthttp.RequestHandler, 0, len(handlers))
	for _, h := range handlers {
		switch handler := h.(type) {
		case fasthttp.RequestHandler:
			result = append(result, handler)
		case func(*fasthttp.Context):
			result = append(result, handler)
		}
	}
	return result
}

// FiberTransportFactory Fiber 传输工厂
type FiberTransportFactory struct{}

// NewFiberTransportFactory 创建 Fiber 传输工厂
func NewFiberTransportFactory() *FiberTransportFactory {
	return &FiberTransportFactory{}
}

// CreateServer 创建 Fiber 服务器
func (f *FiberTransportFactory) CreateServer(name string, opts ...Option) (Server, error) {
	return NewFiberServer(name, opts...), nil
}

// GetType 获取传输类型
func (f *FiberTransportFactory) GetType() TransportType {
	return TransportTypeFiber
}

// FiberRequest Fiber 请求封装
// 提供了对 fasthttp.RequestCtx 的便捷访问
type FiberRequest struct {
	ctx *fasthttp.RequestCtx
}

// NewFiberRequest 创建新的 Fiber 请求
func NewFiberRequest(ctx *fasthttp.RequestCtx) *FiberRequest {
	return &FiberRequest{ctx: ctx}
}

// Method 获取请求方法
func (r *FiberRequest) Method() string {
	return string(r.ctx.Method())
}

// URL 获取请求 URL
func (r *FiberRequest) URL() string {
	return string(r.ctx.Request.URI().FullURI())
}

// Header 获取请求头
func (r *FiberRequest) Header(name string) string {
	return string(r.ctx.Request.Header.Peek(name))
}

// Body 获取请求体
func (r *FiberRequest) Body() []byte {
	return r.ctx.Request.Body()
}

// Context 获取请求上下文
func (r *FiberRequest) Context() context.Context {
	return r.ctx
}

// FiberResponse Fiber 响应封装
// 提供了对 fasthttp.ResponseCtx 的便捷访问
type FiberResponse struct {
	ctx *fasthttp.RequestCtx
}

// NewFiberResponse 创建新的 Fiber 响应
func NewFiberResponse(ctx *fasthttp.RequestCtx) *FiberResponse {
	return &FiberResponse{ctx: ctx}
}

// Status 设置响应状态码
func (r *FiberResponse) Status(code int) {
	r.ctx.Response.SetStatusCode(code)
}

// Header 设置响应头
func (r *FiberResponse) Header(name, value string) {
	r.ctx.Response.Header.Set(name, value)
}

// Write 写入响应体
func (r *FiberResponse) Write(data []byte) error {
	_, err := r.ctx.Write(data)
	return err
}
