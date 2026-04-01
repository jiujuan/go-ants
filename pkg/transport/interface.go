package transport

import (
	"context"
	"net/http"
	"time"
)

// Server HTTP 服务器接口
// 所有 HTTP 服务器实现必须实现此接口
type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
	// Name 获取服务器名称
	Name() string
}

// RouterHandler 路由处理器接口
// 用于定义不同框架的路由处理方式
type RouterHandler interface {
	// GET 注册 GET 路由
	GET(relativePath string, handlers ...interface{})
	// POST 注册 POST 路由
	POST(relativePath string, handlers ...interface{})
	// PUT 注册 PUT 路由
	PUT(relativePath string, handlers ...interface{})
	// DELETE 注册 DELETE 路由
	DELETE(relativePath string, handlers ...interface{})
	// PATCH 注册 PATCH 路由
	PATCH(relativePath string, handlers ...interface{})
	// OPTIONS 注册 OPTIONS 路由
	OPTIONS(relativePath string, handlers ...interface{})
	// HEAD 注册 HEAD 路由
	HEAD(relativePath string, handlers ...interface{})
	// Group 创建路由组
	Group(relativePath string, handlers ...interface{}) interface{}
	// Use 注册中间件
	Use(middleware ...interface{})
}

// Middleware 中间件接口
// 用于定义 HTTP 中间件
type Middleware interface {
	// Name 获取中间件名称
	Name() string
	// Handle 处理请求
	Handle(next http.Handler) http.Handler
}

// MiddlewareFunc 中间件函数类型
// 允许使用函数作为中间件
type MiddlewareFunc func(next http.Handler) http.Handler

// Handle 实现 Middleware 接口
func (m MiddlewareFunc) Handle(next http.Handler) http.Handler {
	return m(next)
}

// TransportType 传输类型
type TransportType string

const (
	// TransportTypeGin Gin 传输类型
	TransportTypeGin TransportType = "gin"
	// TransportTypeFiber Fiber 传输类型
	TransportTypeFiber TransportType = "fiber"
	// TransportTypeMux Mux 传输类型
	TransportTypeMux TransportType = "mux"
)

// TransportFactory 传输工厂接口
// 用于创建不同类型的 HTTP 服务器
type TransportFactory interface {
	// CreateServer 创建服务器实例
	CreateServer(name string, opts ...Option) (Server, error)
	// GetType 获取传输类型
	GetType() TransportType
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Network      string
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Host         string
	Port         int
	Handler      http.Handler
	Middleware   []Middleware
}

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Network:      "tcp",
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}
}

// Request HTTP 请求接口
type Request interface {
	// Method 获取请求方法
	Method() string
	// URL 获取请求 URL
	URL() string
	// Header 获取请求头
	Header(name string) string
	// Body 获取请求体
	Body() []byte
	// Context 获取请求上下文
	Context() context.Context
}

// Response HTTP 响应接口
type Response interface {
	// Status 设置响应状态码
	Status(code int)
	// Header 设置响应头
	Header(name, value string)
	// Write 写入响应体
	Write(data []byte) error
	// JSON 返回 JSON 响应
	JSON(code int, obj interface{}) error
}
