package transport

import (
	"context"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== 标准中间件实现 =====

// CorsMiddleware 跨域中间件
// 允许来自任何源的跨域请求
func CorsMiddleware() Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

// TimeoutMiddleware 超时中间件
// 为请求设置超时限制
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
			}
		})
	})
}

// LoggingMiddleware 日志中间件
// 记录请求的详细信息
func LoggingMiddleware() Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应包装器以捕获状态码
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			latency := time.Since(start)

			log.Info("http request",
				log.String("method", r.Method),
				log.String("path", r.URL.Path),
				log.String("query", r.URL.RawQuery),
				log.Int("status", wrapped.statusCode),
				log.Duration("latency", latency),
				log.String("client_ip", r.RemoteAddr))
		})
	})
}

// responseWriter 响应包装器
// 用于捕获 HTTP 响应状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 捕获状态码
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoveryMiddleware 恢复中间件
// 捕获 panic 并返回 500 错误
func RecoveryMiddleware() Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						log.Any("error", err),
						log.String("stack", string(debug.Stack())))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	})
}

// RequestIDMiddleware 请求ID中间件
// 为每个请求生成唯一的 ID
func RequestIDMiddleware() Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
}

// generateRequestID 生成唯一的请求 ID
func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}

// ===== Gin 中间件 =====

// GinCorsMiddleware Gin CORS 中间件
func GinCorsMiddleware() interface{} {
	return func(c interface{}) {
		if gin, ok := c.(interface{ Header(string, string) }); ok {
			gin.Header("Access-Control-Allow-Origin", "*")
		}
	}
}

// GinTimeoutMiddleware Gin 超时中间件
func GinTimeoutMiddleware(timeout time.Duration) interface{} {
	return func(c interface{}) {
		// Gin 的超时处理通过 context 实现
		// 这里提供一个基础实现
	}
}

// ===== Fiber 中间件 =====

// FiberCorsMiddleware Fiber CORS 中间件
func FiberCorsMiddleware() interface{} {
	return func(c interface{}) {
		// Fiber 的 CORS 处理
	}
}

// FiberTimeoutMiddleware Fiber 超时中间件
func FiberTimeoutMiddleware(timeout time.Duration) interface{} {
	return func(c interface{}) {
		// Fiber 的超时处理
	}
}

// ===== 中间件工具函数 =====

// ToMiddleware 将函数转换为 Middleware
func ToMiddleware(f func(http.Handler) http.Handler) Middleware {
	return MiddlewareFunc(f)
}

// Chain 中间件链
// 允许将多个中间件链接在一起
type Chain struct {
	middlewares []Middleware
}

// NewChain 创建新的中间件链
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Then 执行中间件链
func (c *Chain) Then(h http.Handler) http.Handler {
	if len(c.middlewares) == 0 {
		return h
	}

	// 从最后一个中间件开始
	result := h
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		result = c.middlewares[i].Handle(result)
	}

	return result
}

// Append 添加中间件到链尾
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Prepend 添加中间件到链头
func (c *Chain) Prepend(middlewares ...Middleware) *Chain {
	c.middlewares = append(middlewares, c.middlewares...)
	return c
}

// MiddlewareName 获取中间件名称
type MiddlewareName interface {
	Name() string
}

// GetMiddlewareName 获取中间件的名称
func GetMiddlewareName(m Middleware) string {
	return m.Name()
}

// DefaultMiddlewareName 默认中间件名称
type DefaultMiddlewareName struct {
	name string
}

// Name 获取名称
func (m *DefaultMiddlewareName) Name() string {
	return m.name
}

// WithName 创建带名称的中间件
func WithName(name string, f func(http.Handler) http.Handler) Middleware {
	return &namedMiddleware{
		name:  name,
		func_: f,
	}
}

// namedMiddleware 带名称的中间件
type namedMiddleware struct {
	name  string
	func_ func(http.Handler) http.Handler
}

// Handle 实现 Middleware 接口
func (m *namedMiddleware) Handle(next http.Handler) http.Handler {
	return m.func_(next)
}

// Name 实现 MiddlewareName 接口
func (m *namedMiddleware) Name() string {
	return m.name
}

// MiddlewareFuncName 获取中间件函数的名称
func MiddlewareFuncName(f func(http.Handler) http.Handler) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
