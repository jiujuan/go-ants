package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/valyala/fasthttp"

	"github.com/jiujuan/go-ants/pkg/log"
)

// Server HTTP 服务器接口
type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
	// Name 获取服务器名称
	Name() string
}

// Option 是服务器选项函数
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
func WithNetwork(network string) Option {
	return func(o *Options) {
		o.Network = network
	}
}

// WithAddr 设置地址
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithReadTimeout 设置读取超时
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时
func WithIdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// WithHost 设置主机
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort 设置端口
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithHandler 设置处理器
func WithHandler(handler http.Handler) Option {
	return func(o *Options) {
		o.Handler = handler
	}
}

// WithMiddleware 设置中间件
func WithMiddleware(middleware ...Middleware) Option {
	return func(o *Options) {
		o.Middleware = append(o.Middleware, middleware...)
	}
}

// ===== Gin 服务器 =====

// GinServer Gin HTTP 服务器
type GinServer struct {
	name         string
	engine       *gin.Engine
	httpServer   *http.Server
	opts         *Options
	middlewares  []gin.HandlerFunc
	RouterGroups []gin.RouterGroup
}

// NewGinServer 创建新的 Gin 服务器
func NewGinServer(name string, opts ...Option) *GinServer {
	options := &Options{
		Network:      "tcp",
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	for _, opt := range opts {
		opt(options)
	}

	// 设置 Gin 模式
	if mode := log.GetGlobalLogger().Core().(type); mode != nil {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// 添加全局中间件
	globalMiddlewares := []gin.HandlerFunc{
		Recovery(),
		Logger(),
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

// Start 启动服务器
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

// Stop 停止服务器
func (s *GinServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("gin server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// Engine 获取 Gin 引擎
func (s *GinServer) Engine() *gin.Engine {
	return s.engine
}

// HttpServer 获取 HTTP 服务器
func (s *GinServer) HttpServer() *http.Server {
	return s.httpServer
}

// GET 注册 GET 路由
func (s *GinServer) GET(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.GET(relativePath, handlers...)
}

// POST 注册 POST 路由
func (s *GinServer) POST(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.POST(relativePath, handlers...)
}

// PUT 注册 PUT 路由
func (s *GinServer) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.PUT(relativePath, handlers...)
}

// DELETE 注册 DELETE 路由
func (s *GinServer) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.DELETE(relativePath, handlers...)
}

// PATCH 注册 PATCH 路由
func (s *GinServer) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.PATCH(relativePath, handlers...)
}

// OPTIONS 注册 OPTIONS 路由
func (s *GinServer) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.OPTIONS(relativePath, handlers...)
}

// HEAD 注册 HEAD 路由
func (s *GinServer) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	s.engine.HEAD(relativePath, handlers...)
}

// Group 创建路由组
func (s *GinServer) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return s.engine.Group(relativePath, handlers...)
}

// Use 注册中间件
func (s *GinServer) Use(middleware ...gin.HandlerFunc) {
	s.engine.Use(middleware...)
}

// ===== Fiber 服务器 =====

// FiberServer Fiber HTTP 服务器
type FiberServer struct {
	name        string
	app         *fasthttp.App
	httpServer  *http.Server
	opts        *Options
	middlewares []fasthttp.RequestHandler
}

// NewFiberServer 创建新的 Fiber 服务器
func NewFiberServer(name string, opts ...Option) *FiberServer {
	options := &Options{
		Network:      "tcp",
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	for _, opt := range opts {
		opt(options)
	}

	app := fasthttp.New()

	server := &FiberServer{
		name:       name,
		app:        app,
		opts:       options,
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

// Start 启动服务器
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

// Stop 停止服务器
func (s *FiberServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("fiber server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// App 获取 Fiber 应用
func (s *FiberServer) App() *fasthttp.App {
	return s.app
}

// ===== Mux 服务器 =====

// MuxServer Mux HTTP 服务器
type MuxServer struct {
	name        string
	router      *mux.Router
	httpServer  *http.Server
	opts        *Options
	middlewares []mux.MiddlewareFunc
}

// NewMuxServer 创建新的 Mux 服务器
func NewMuxServer(name string, opts ...Option) *MuxServer {
	options := &Options{
		Network:      "tcp",
		Addr:         ":8080",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

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

// Start 启动服务器
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

// Stop 停止服务器
func (s *MuxServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Info("mux server stopping", log.String("name", s.name))
	return s.httpServer.Shutdown(ctx)
}

// Router 获取 Mux 路由器
func (s *MuxServer) Router() *mux.Router {
	return s.router
}

// ===== 中间件相关 =====

// Middleware 中间件接口
type Middleware interface {
	Name() string
	Handle(next http.Handler) http.Handler
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(next http.Handler) http.Handler

// Handle 实现 Middleware 接口
func (m MiddlewareFunc) Handle(next http.Handler) http.Handler {
	return m(next)
}

// Name 实现 Name 方法
func (m MiddlewareFunc) Name() string {
	return runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					log.Error(err.(error)),
					log.String("stack", string(debug.Stack())))
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
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

// ToGinMiddleware 将自定义中间件转换为 Gin 中间件
func ToGinMiddleware(m Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

// CorsMiddleware 跨域中间件
func CorsMiddleware() Middleware {
	return MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

// TimeoutMiddleware 超时中间件
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
				w.WriteHeader(http.StatusRequestTimeout)
				w.Write([]byte("Request Timeout"))
			}
		})
	})
}
