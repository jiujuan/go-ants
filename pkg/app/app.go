// Package ants 是 go-ants 框架的核心入口，提供了应用程序生命周期管理和依赖注入工具。
package ants

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jiujuan/go-ants/pkg/log"
)

// App represents the application instance.
// App 代表应用程序实例，提供生命周期管理功能。
type App struct {
	ctx       context.Context
	cancel    context.CancelFunc
	opts      *options
	component []Component
}

// Component 定义了应用程序组件的接口
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Option 是配置 App 的函数选项
type Option func(*options)

type options struct {
	name       string
	ctx        context.Context
	sigs       []os.Signal
	waitTime   time.Duration
	logger     *log.Logger
	components []Component
}

// WithName 设置应用名称
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithSignal 设置信号量
func WithSignal(sigs []os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

// WithWaitTime 设置等待时间
func WithWaitTime(waitTime time.Duration) Option {
	return func(o *options) {
		o.waitTime = waitTime
	}
}

// WithLogger 设置日志实例
func WithLogger(logger *log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithComponents 设置组件列表
func WithComponents(components ...Component) Option {
	return func(o *options) {
		o.components = components
	}
}

// New 创建一个新的应用实例
func New(opts ...Option) (*App, func()) {
	options := &options{
		name:     "go-ants",
		ctx:      context.Background(),
		sigs:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		waitTime: time.Second * 10,
		logger:   log.DefaultLogger(),
	}

	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithCancel(options.ctx)

	app := &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   options,
	}

	cleanup := func() {
		cancel()
	}

	return app, cleanup
}

// Run 启动应用程序
func (a *App) Run() error {
	a.opts.logger.Info(a.opts.name, log.String("msg", "application starting"))

	// 启动所有组件
	for _, component := range a.opts.components {
		if err := component.Start(a.ctx); err != nil {
			a.opts.logger.Error("component start failed",
				log.String("component", getComponentName(component)),
				log.Error(err))
			return err
		}
	}

	// 等待退出信号
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, a.opts.sigs...)

	select {
	case <-a.ctx.Done():
		a.opts.logger.Info(a.opts.name, log.String("msg", "context cancelled"))
	case sig := <-stopCh:
		a.opts.logger.Info(a.opts.name,
			log.String("msg", "received signal"),
			log.String("signal", sig.String()))
	}

	// 停止所有组件
	a.opts.logger.Info(a.opts.name, log.String("msg", "shutting down..."))
	stopCtx, stopCancel := context.WithTimeout(context.Background(), a.opts.waitTime)
	defer stopCancel()

	for _, component := range a.opts.components {
		if err := component.Stop(stopCtx); err != nil {
			a.opts.logger.Error("component stop failed",
				log.String("component", getComponentName(component)),
				log.Error(err))
		}
	}

	a.opts.logger.Info(a.opts.name, log.String("msg", "application stopped"))
	return nil
}

// Context 返回应用的上下文
func (a *App) Context() context.Context {
	return a.ctx
}

func getComponentName(c Component) string {
	if named, ok := c.(interface{ Name() string }); ok {
		return named.Name()
	}
	return "unknown"
}

// NamedComponent 命名组件接口
type NamedComponent interface {
	Name() string
	Component
}
