// Package server 提供了 HTTP 服务器的初始化和配置。
package server

import (
	"context"
	"net/http"

	"github.com/google/wire"
	"github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/transport"
)

// HTTPServer HTTP 服务器包装器
type HTTPServer struct {
	name       string
	server     transport.Server
	engine     interface{}
	middleware []interface{}
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(name string, server transport.Server) *HTTPServer {
	return &HTTPServer{
		name:   name,
		server: server,
	}
}

// Name 获取服务器名称
func (s *HTTPServer) Name() string {
	return s.name
}

// Start 启动服务器
func (s *HTTPServer) Start(ctx context.Context) error {
	log.Info("http server starting",
		log.String("name", s.name))

	return s.server.Start()
}

// Stop 停止服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	log.Info("http server stopping",
		log.String("name", s.name))

	return s.server.Stop()
}

// Engine 获取引擎（Gin 或 Fiber）
func (s *HTTPServer) Engine() interface{} {
	return s.engine
}

// SetEngine 设置引擎
func (s *HTTPServer) SetEngine(engine interface{}) {
	s.engine = engine
}

// ===== 服务器构建器 =====

// HTTPServerBuilder HTTP 服务器构建器
type HTTPServerBuilder struct {
	name       string
	address    string
	middleware []interface{}
}

// NewHTTPServerBuilder 创建服务器构建器
func NewHTTPServerBuilder() *HTTPServerBuilder {
	return &HTTPServerBuilder{
		name:    "http",
		address: ":8080",
	}
}

// WithName 设置名称
func (b *HTTPServerBuilder) WithName(name string) *HTTPServerBuilder {
	b.name = name
	return b
}

// WithAddress 设置地址
func (b *HTTPServerBuilder) WithAddress(addr string) *HTTPServerBuilder {
	b.address = addr
	return b
}

// WithMiddleware 添加中间件
func (b *HTTPServerBuilder) WithMiddleware(m ...interface{}) *HTTPServerBuilder {
	b.middleware = append(b.middleware, m...)
	return b
}

// BuildGin 构建 Gin 服务器
func (b *HTTPServerBuilder) BuildGin() *HTTPServer {
	server := transport.NewGinServer(b.name,
		transport.WithAddr(b.address),
	)

	return NewHTTPServer(b.name, server)
}

// BuildFiber 构建 Fiber 服务器
func (b *HTTPServerBuilder) BuildFiber() *HTTPServer {
	server := transport.NewFiberServer(b.name,
		transport.WithAddr(b.address),
	)

	return NewHTTPServer(b.name, server)
}

// ===== 应用服务器 =====

// Server 应用服务器
type Server struct {
	httpserver *HTTPServer
}

// New 创建应用服务器
func New(httpServer *HTTPServer) *Server {
	return &Server{
		httpserver: httpServer,
	}
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	return s.httpserver.Start(ctx)
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.httpserver.Stop(ctx)
}

// Name 获取名称
func (s *Server) Name() string {
	return "server"
}

// ProviderSet Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGinServer,
	NewFiberServer,
	NewServer,
)

// NewGinServer 创建 Gin 服务器
func NewGinServer(address string) *HTTPServer {
	server := transport.NewGinServer("gin",
		transport.WithAddr(address),
	)
	return NewHTTPServer("gin", server)
}

// NewFiberServer 创建 Fiber 服务器
func NewFiberServer(address string) *HTTPServer {
	server := transport.NewFiberServer("fiber",
		transport.WithAddr(address),
	)
	return NewHTTPServer("fiber", server)
}

// 确保 HTTPServer 实现 app.Component 接口
var _ app.Component = (*HTTPServer)(nil)
var _ app.NamedComponent = (*HTTPServer)(nil)
