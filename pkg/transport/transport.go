// Package transport provides HTTP server implementations using Gin and Fiber.
// Package transport 提供了基于 Gin 和 Fiber 的 HTTP 服务器实现。
//
// 该模块已重构为模块化设计，支持多种 HTTP 框架：
//   - Gin: 高性能的 Web 框架
//   - Fiber: 最快的 Go HTTP 框架
//   - Mux: 强大的路由匹配和变量支持
//
// 重构后的文件结构：
//   - interface.go: 公共接口定义
//   - options.go: 配置选项
//   - gin.go: Gin 服务器实现
//   - fiber.go: Fiber 服务器实现
//   - mux.go: Mux 服务器实现
//   - middleware.go: 中间件实现
//   - transport.go: 向后兼容入口文件
package transport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ===== 导出接口和类型 =====

// Server HTTP 服务器接口
// 所有 HTTP 服务器实现必须实现此接口
//
// 示例:
//
//	var _ Server = (*GinServer)(nil)
type Server = Server

// RouterHandler 路由处理器接口
type RouterHandler = RouterHandler

// Middleware 中间件接口
type Middleware = Middleware

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc = MiddlewareFunc

// TransportType 传输类型
type TransportType = TransportType

// TransportFactory 传输工厂接口
type TransportFactory = TransportFactory

// ServerConfig 服务器配置
type ServerConfig = ServerConfig

// Options 服务器选项
type Options = Options

// Option 服务器选项函数
type Option = Option

// ===== 导出常量 =====

const (
	// TransportTypeGin Gin 传输类型
	TransportTypeGin = TransportTypeGin
	// TransportTypeFiber Fiber 传输类型
	TransportTypeFiber = TransportTypeFiber
	// TransportTypeMux Mux 传输类型
	TransportTypeMux = TransportTypeMux
)

// ===== 导出 Gin 服务器 =====

// GinServer Gin HTTP 服务器实现
type GinServer = GinServer

// NewGinServer 创建新的 Gin 服务器实例
func NewGinServer(name string, opts ...Option) *GinServer {
	return NewGinServer(name, opts...)
}

// GinRecovery Gin 恢复中间件
func GinRecovery() gin.HandlerFunc {
	return GinRecovery()
}

// GinLogger Gin 日志中间件
func GinLogger() gin.HandlerFunc {
	return GinLogger()
}

// ToGinMiddleware 将标准中间件转换为 Gin 中间件
func ToGinMiddleware(m Middleware) gin.HandlerFunc {
	return ToGinMiddleware(m)
}

// GinTransportFactory Gin 传输工厂
type GinTransportFactory = GinTransportFactory

// NewGinTransportFactory 创建 Gin 传输工厂
func NewGinTransportFactory() *GinTransportFactory {
	return NewGinTransportFactory()
}

// ===== 导出 Fiber 服务器 =====

// FiberServer Fiber HTTP 服务器实现
type FiberServer = FiberServer

// NewFiberServer 创建新的 Fiber 服务器实例
func NewFiberServer(name string, opts ...Option) *FiberServer {
	return NewFiberServer(name, opts...)
}

// FiberTransportFactory Fiber 传输工厂
type FiberTransportFactory = FiberTransportFactory

// NewFiberTransportFactory 创建 Fiber 传输工厂
func NewFiberTransportFactory() *FiberTransportFactory {
	return NewFiberTransportFactory()
}

// FiberRequest Fiber 请求封装
type FiberRequest = FiberRequest

// NewFiberRequest 创建新的 Fiber 请求
func NewFiberRequest(ctx interface{}) *FiberRequest {
	return &FiberRequest{}
}

// FiberResponse Fiber 响应封装
type FiberResponse = FiberResponse

// NewFiberResponse 创建新的 Fiber 响应
func NewFiberResponse(ctx interface{}) *FiberResponse {
	return &FiberResponse{}
}

// ===== 导出 Mux 服务器 =====

// MuxServer Mux HTTP 服务器实现
type MuxServer = MuxServer

// NewMuxServer 创建新的 Mux 服务器实例
func NewMuxServer(name string, opts ...Option) *MuxServer {
	return NewMuxServer(name, opts...)
}

// MuxTransportFactory Mux 传输工厂
type MuxTransportFactory = MuxTransportFactory

// NewMuxTransportFactory 创建 Mux 传输工厂
func NewMuxTransportFactory() *MuxTransportFactory {
	return NewMuxTransportFactory()
}

// MuxRequest Mux 请求封装
type MuxRequest = MuxRequest

// NewMuxRequest 创建新的 Mux 请求
func NewMuxRequest(r *http.Request) *MuxRequest {
	return NewMuxRequest(r)
}

// MuxResponse Mux 响应封装
type MuxResponse = MuxResponse

// NewMuxResponse 创建新的 Mux 响应
func NewMuxResponse(w http.ResponseWriter) *MuxResponse {
	return NewMuxResponse(w)
}

// ===== 导出选项函数 =====

// WithNetwork 设置网络类型
func WithNetwork(network string) Option {
	return WithNetwork(network)
}

// WithAddr 设置地址
func WithAddr(addr string) Option {
	return WithAddr(addr)
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(timeout time.Duration) Option {
	return WithReadTimeout(timeout)
}

// WithWriteTimeout 设置写入超时时间
func WithWriteTimeout(timeout time.Duration) Option {
	return WithWriteTimeout(timeout)
}

// WithIdleTimeout 设置空闲超时时间
func WithIdleTimeout(timeout time.Duration) Option {
	return WithIdleTimeout(timeout)
}

// WithHost 设置主机名
func WithHost(host string) Option {
	return WithHost(host)
}

// WithPort 设置端口号
func WithPort(port int) Option {
	return WithPort(port)
}

// WithHandler 设置 HTTP 处理器
func WithHandler(handler http.Handler) Option {
	return WithHandler(handler)
}

// WithMiddleware 设置中间件
func WithMiddleware(middleware ...Middleware) Option {
	return WithMiddleware(middleware...)
}

// WithMiddlewareFunc 设置函数类型的中间件
func WithMiddlewareFunc(middlewareFunc ...func(http.Handler) http.Handler) Option {
	return func(o *Options) {}
}

// MergeOptions 合并多个选项
func MergeOptions(opts ...Option) Option {
	return MergeOptions(opts...)
}

// ApplyOptions 将选项应用到 Options 结构体
func ApplyOptions(options *Options, opts ...Option) {
	ApplyOptions(options, opts...)
}

// NewOptions 创建默认选项
func NewOptions() *Options {
	return NewOptions()
}

// ===== 导出中间件 =====

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return GinRecovery()
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return GinLogger()
}

// CorsMiddleware 跨域中间件
func CorsMiddleware() Middleware {
	return CorsMiddleware()
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return TimeoutMiddleware(timeout)
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return LoggingMiddleware()
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() Middleware {
	return RecoveryMiddleware()
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() Middleware {
	return RequestIDMiddleware()
}

// Chain 中间件链
type Chain = Chain

// NewChain 创建新的中间件链
func NewChain(middlewares ...Middleware) *Chain {
	return NewChain(middlewares...)
}

// ToMiddleware 将函数转换为 Middleware
func ToMiddleware(f func(http.Handler) http.Handler) Middleware {
	return ToMiddleware(f)
}

// WithName 创建带名称的中间件
func WithName(name string, f func(http.Handler) http.Handler) Middleware {
	return WithName(name, f)
}

// ===== 导出工具函数 =====

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
	return DefaultServerConfig()
}
