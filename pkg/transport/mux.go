package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/jiujuan/go-ants/pkg/log"
)

// MuxServer Mux HTTP 服务器实现
// 基于 Gorilla Mux 的 HTTP 路由器
// Mux 提供了强大的路由匹配和变量支持
type MuxServer struct {
	name        string
	router      *mux.Router
	httpServer  *http.Server
	opts        *Options
	middlewares []mux.MiddlewareFunc
}

// NewMuxServer 创建新的 Mux 服务器实例
// name: 服务器名称
// opts: 可选的配置选项
func NewMuxServer(name string, opts ...Option) *MuxServer {
	options := NewOptions()

	for _, opt := range opts {
		opt(options)
	}

	router := mux.NewRouter()

	server := &MuxServer{
		name:   name,
		router: router,
		opts:   options,
	}

	// 创建 HTTP 服务器
	server.httpServer = &http.Server{
		Addr:         options.Addr,
		Handler:      router,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		IdleTimeout:  options.IdleTimeout,
	}

	return server
}

// Name 获取服务器名称
func (s *MuxServer) Name() string {
	return s.name
}

// Start 启动 Mux 服务器
func (s *MuxServer) Start() error {
	ln, err := net.Listen(s.opts.Network, s.opts.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Info("mux server starting",
		log.String("name", s.name),
		log.String("addr", s.opts.Addr))

	return s.httpServer.Serve(ln)
}

// Stop 停止 Mux 服务器
func (s *MuxServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("mux server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// Router 获取 Mux 路由器实例
// 返回 *mux.Router，可用于直接访问 Mux 的高级功能
func (s *MuxServer) Router() *mux.Router {
	return s.router
}

// HttpServer 获取底层 HTTP 服务器
// 返回 *http.Server，可用于获取服务器配置信息
func (s *MuxServer) HttpServer() *http.Server {
	return s.httpServer
}

// GET 注册 GET 路由
func (s *MuxServer) GET(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("GET")
}

// POST 注册 POST 路由
func (s *MuxServer) POST(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("POST")
}

// PUT 注册 PUT 路由
func (s *MuxServer) PUT(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("PUT")
}

// DELETE 注册 DELETE 路由
func (s *MuxServer) DELETE(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("DELETE")
}

// PATCH 注册 PATCH 路由
func (s *MuxServer) PATCH(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("PATCH")
}

// OPTIONS 注册 OPTIONS 路由
func (s *MuxServer) OPTIONS(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("OPTIONS")
}

// HEAD 注册 HEAD 路由
func (s *MuxServer) HEAD(relativePath string, handlers ...interface{}) {
	httpHandlers := s.convertHandlers(handlers)
	s.router.Handle(relativePath, httpHandlers...).Methods("HEAD")
}

// Group 创建路由组 (Mux 中使用 StrictSlash)
func (s *MuxServer) Group(relativePath string, handlers ...interface{}) interface{} {
	return s.router.PathPrefix(relativePath)
}

// Use 注册中间件
func (s *MuxServer) Use(middleware ...interface{}) {
	for _, m := range middleware {
		switch mw := m.(type) {
		case mux.MiddlewareFunc:
			s.router.Use(mw)
		case func(http.Handler) http.Handler:
			s.router.Use(mw)
		}
	}
}

// convertHandlers 将通用接口转换为 http.Handler
func (s *MuxServer) convertHandlers(handlers []interface{}) http.Handler {
	if len(handlers) == 0 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}

	if len(handlers) == 1 {
		return s.toHandler(handlers[0])
	}

	// 将多个处理器链接成一个
	chain := make([]func(http.Handler) http.Handler, len(handlers))
	for i, h := range handlers {
		chain[i] = func(h interface{}) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return s.toHandler(h)
			}
		}(h)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		for i := len(chain) - 1; i >= 0; i-- {
			handler = chain[i](handler)
		}
		handler.ServeHTTP(w, r)
	})
}

// toHandler 将通用接口转换为 http.Handler
func (s *MuxServer) toHandler(h interface{}) http.Handler {
	switch handler := h.(type) {
	case http.Handler:
		return handler
	case func(http.ResponseWriter, *http.Request):
		return http.HandlerFunc(handler)
	default:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}
}

// MuxTransportFactory Mux 传输工厂
type MuxTransportFactory struct{}

// NewMuxTransportFactory 创建 Mux 传输工厂
func NewMuxTransportFactory() *MuxTransportFactory {
	return &MuxTransportFactory{}
}

// CreateServer 创建 Mux 服务器
func (f *MuxTransportFactory) CreateServer(name string, opts ...Option) (Server, error) {
	return NewMuxServer(name, opts...), nil
}

// GetType 获取传输类型
func (f *MuxTransportFactory) GetType() TransportType {
	return TransportTypeMux
}

// MuxRequest Mux 请求封装
// 提供了对 http.Request 的便捷访问
type MuxRequest struct {
	request *http.Request
}

// NewMuxRequest 创建新的 Mux 请求
func NewMuxRequest(r *http.Request) *MuxRequest {
	return &MuxRequest{request: r}
}

// Method 获取请求方法
func (r *MuxRequest) Method() string {
	return r.request.Method
}

// URL 获取请求 URL
func (r *MuxRequest) URL() string {
	return r.request.URL.String()
}

// Header 获取请求头
func (r *MuxRequest) Header(name string) string {
	return r.request.Header.Get(name)
}

// Body 获取请求体
func (r *MuxRequest) Body() []byte {
	if r.request.Body == nil {
		return nil
	}
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, _ := r.request.Body.Read(buf)
		if n == 0 {
			break
		}
		body = append(body, buf[:n]...)
	}
	return body
}

// Context 获取请求上下文
func (r *MuxRequest) Context() context.Context {
	return r.request.Context()
}

// MuxResponse Mux 响应封装
// 提供了对 http.ResponseWriter 的便捷访问
type MuxResponse struct {
	writer http.ResponseWriter
}

// NewMuxResponse 创建新的 Mux 响应
func NewMuxResponse(w http.ResponseWriter) *MuxResponse {
	return &MuxResponse{writer: w}
}

// Status 设置响应状态码
func (r *MuxResponse) Status(code int) {
	r.writer.WriteHeader(code)
}

// Header 设置响应头
func (r *MuxResponse) Header(name, value string) {
	r.writer.Header().Set(name, value)
}

// Write 写入响应体
func (r *MuxResponse) Write(data []byte) error {
	_, err := r.writer.Write(data)
	return err
}
