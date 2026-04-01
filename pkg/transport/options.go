package transport

import (
	"net/http"
	"time"
)

// Option 是服务器选项函数类型
// 用于以函数选项模式配置服务器
type Option func(*Options)

type Options struct {
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

// WithNetwork 设置网络类型
// network: 网络类型，如 "tcp", "udp" 等
func WithNetwork(network string) Option {
	return func(o *Options) {
		o.Network = network
	}
}

// WithAddr 设置地址
// addr: 服务器监听地址，如 ":8080", "localhost:8080"
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithReadTimeout 设置读取超时时间
// timeout: 读取超时时长
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时时间
// timeout: 写入超时时长
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时时间
// timeout: 空闲超时时长
func WithIdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// WithHost 设置主机名
// host: 主机名或 IP 地址
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort 设置端口号
// port: 端口号
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithHandler 设置 HTTP 处理器
// handler: http.Handler 实例
func WithHandler(handler http.Handler) Option {
	return func(o *Options) {
		o.Handler = handler
	}
}

// WithMiddleware 设置中间件
// middleware: 可变数量的中间件
func WithMiddleware(middleware ...Middleware) Option {
	return func(o *Options) {
		o.Middleware = append(o.Middleware, middleware...)
	}
}

// WithMiddlewareFunc 设置函数类型的中间件
// middlewareFunc: 可变数量的中间件函数
func WithMiddlewareFunc(middlewareFunc ...MiddlewareFunc) Option {
	return func(o *Options) {
		for _, m := range middlewareFunc {
			o.Middleware = append(o.Middleware, m)
		}
	}
}

// MergeOptions 合并多个选项
// opts: 可变数量的选项
func MergeOptions(opts ...Option) Option {
	return func(o *Options) {
		for _, opt := range opts {
			opt(o)
		}
	}
}

// ApplyOptions 将选项应用到 Options 结构体
// options: 目标 Options 结构体指针
// opts: 要应用的选项
func ApplyOptions(options *Options, opts ...Option) {
	for _, opt := range opts {
		opt(options)
	}
}

// NewOptions 创建默认选项
func NewOptions() *Options {
	return &Options{
		Network:      "tcp",
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}
}
