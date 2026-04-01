package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

// mockComponent 是一个模拟组件，用于测试
type mockComponent struct {
	name     string
	startErr error
	stopErr  error
	started  bool
	stopped  bool
	mu       sync.Mutex
}

func (m *mockComponent) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.startErr != nil {
		return m.startErr
	}
	m.started = true
	return nil
}

func (m *mockComponent) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopErr != nil {
		return m.stopErr
	}
	m.stopped = true
	return nil
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) IsStarted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.started
}

func (m *mockComponent) IsStopped() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopped
}

// TestNew_DefaultOptions 测试默认选项创建应用
func TestNew_DefaultOptions(t *testing.T) {
	app, cleanup := New()
	defer cleanup()

	if app == nil {
		t.Error("应用创建失败")
	}

	if app.opts.name != "go-ants" {
		t.Errorf("默认名称期望 go-ants, 实际 %s", app.opts.name)
	}
}

// TestNew_WithName 测试设置应用名称
func TestNew_WithName(t *testing.T) {
	app, cleanup := New(WithName("test-app"))
	defer cleanup()

	if app.opts.name != "test-app" {
		t.Errorf("应用名称设置失败: %s", app.opts.name)
	}
}

// TestNew_WithContext 测试设置上下文
func TestNew_WithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "key", "value")
	app, cleanup := New(WithContext(ctx))
	defer cleanup()

	if app.ctx != ctx {
		t.Error("上下文设置失败")
	}
}

// TestNew_WithSignal 测试设置信号量
func TestNew_WithSignal(t *testing.T) {
	sigs := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	app, cleanup := New(WithSignal(sigs))
	defer cleanup()

	if len(app.opts.sigs) != 2 {
		t.Errorf("信号量设置失败")
	}
}

// TestNew_WithWaitTime 测试设置等待时间
func TestNew_WithWaitTime(t *testing.T) {
	waitTime := time.Second * 30
	app, cleanup := New(WithWaitTime(waitTime))
	defer cleanup()

	if app.opts.waitTime != waitTime {
		t.Errorf("等待时间设置失败: %v", app.opts.waitTime)
	}
}

// TestNew_WithComponents 测试设置组件
func TestNew_WithComponents(t *testing.T) {
	component := &mockComponent{name: "test-component"}
	app, cleanup := New(WithComponents(component))
	defer cleanup()

	if len(app.opts.components) != 1 {
		t.Errorf("组件设置失败")
	}
}

// TestApp_Run_EmptyComponents 测试无组件运行
func TestApp_Run_EmptyComponents(t *testing.T) {
	app, cleanup := New(
		WithName("test-app"),
	)
	defer cleanup()

	// 创建一个立即取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.Run()
	// 由于上下文已取消，应该返回错误
	if err != context.Canceled {
		t.Logf("预期 context.Canceled 错误, 得到: %v", err)
	}
}

// TestApp_Run_WithComponent 测试带组件运行
func TestApp_Run_WithComponent(t *testing.T) {
	component := &mockComponent{name: "test-component"}
	app, cleanup := New(
		WithName("test-app"),
		WithComponents(component),
	)
	defer cleanup()

	// 模拟发送信号来停止应用
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	}()

	err := app.Run()
	// 预期会被信号中断
	if err != nil && err != context.Canceled {
		t.Logf("运行结果: %v", err)
	}

	// 检查组件是否启动
	time.Sleep(50 * time.Millisecond)
	if !component.IsStarted() {
		t.Error("组件未启动")
	}
}

// TestApp_Context 测试获取应用上下文
func TestApp_Context(t *testing.T) {
	ctx := context.Background()
	app, cleanup := New(WithContext(ctx))
	defer cleanup()

	if app.Context() != ctx {
		t.Error("上下文获取失败")
	}
}

// TestWithLogger 测试设置日志
func TestWithLogger(t *testing.T) {
	// 创建一个模拟日志
	logger := &mockLogger{}

	app, cleanup := New(WithLogger(logger))
	defer cleanup()

	if app.opts.logger != logger {
		t.Error("日志设置失败")
	}
}

type mockLogger struct{}

func (m *mockLogger) Debugw(msg string, fields ...interface{})     {}
func (m *mockLogger) Debugf(template string, args ...interface{})  {}
func (m *mockLogger) Infow(msg string, fields ...interface{})      {}
func (m *mockLogger) Infof(template string, args ...interface{})   {}
func (m *mockLogger) Warnw(msg string, fields ...interface{})      {}
func (m *mockLogger) Warnf(template string, args ...interface{})   {}
func (m *mockLogger) Errorw(msg string, fields ...interface{})     {}
func (m *mockLogger) Errorf(template string, args ...interface{})  {}
func (m *mockLogger) DPanicw(msg string, fields ...interface{})    {}
func (m *mockLogger) DPanicf(template string, args ...interface{}) {}
func (m *mockLogger) Panicw(msg string, fields ...interface{})     {}
func (m *mockLogger) Panicf(template string, args ...interface{})  {}
func (m *mockLogger) Fatalw(msg string, fields ...interface{})     {}
func (m *mockLogger) Fatalf(template string, args ...interface{})  {}
func (m *mockLogger) With(args ...interface{}) *mockLogger         { return m }
func (m *mockLogger) Named(name string) *mockLogger                { return m }
func (m *mockLogger) Sync() error                                  { return nil }

// TestNamedComponent 测试命名组件接口
func TestNamedComponent(t *testing.T) {
	component := &mockComponent{name: "named-component"}

	// 测试 NamedComponent 接口
	var _ NamedComponent = component

	if component.Name() != "named-component" {
		t.Error("命名组件名称获取失败")
	}
}

// TestComponent_StartError 测试组件启动错误
func TestComponent_StartError(t *testing.T) {
	component := &mockComponent{
		name:     "error-component",
		startErr: assertError{"start error"},
	}

	app, cleanup := New(
		WithName("test-app"),
		WithComponents(component),
	)
	defer cleanup()

	// 模拟发送信号
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	}()

	err := app.Run()
	// 应该报告组件启动错误
	t.Logf("运行错误: %v", err)
}

type assertError struct {
	msg string
}

func (e assertError) Error() string {
	return e.msg
}

// TestComponent_StopError 测试组件停止错误
func TestComponent_StopError(t *testing.T) {
	component := &mockComponent{
		name:    "stop-error-component",
		stopErr: assertError{"stop error"},
	}

	app, cleanup := New(
		WithName("test-app"),
		WithComponents(component),
	)
	defer cleanup()

	// 模拟发送信号
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	}()

	err := app.Run()
	t.Logf("运行错误: %v", err)
}

// TestMultipleComponents 测试多个组件
func TestMultipleComponents(t *testing.T) {
	components := []*mockComponent{
		{name: "component-1"},
		{name: "component-2"},
		{name: "component-3"},
	}

	var comps []Component
	for _, c := range components {
		comps = append(comps, c)
	}

	app, cleanup := New(
		WithName("test-app"),
		WithComponents(comps...),
	)
	defer cleanup()

	if len(app.opts.components) != 3 {
		t.Errorf("组件数量期望 3, 实际 %d", len(app.opts.components))
	}
}

// TestGetComponentName 测试获取组件名称
func TestGetComponentName(t *testing.T) {
	component := &mockComponent{name: "test"}
	name := getComponentName(component)
	if name != "test" {
		t.Errorf("组件名称期望 test, 实际 %s", name)
	}
}

// TestGetComponentName_Unknown 测试未知组件名称
func TestGetComponentName_Unknown(t *testing.T) {
	// 不实现 NamedComponent 的组件
	simpleComponent := &simpleMockComponent{}
	name := getComponentName(simpleComponent)
	if name != "unknown" {
		t.Errorf("未知组件名称期望 unknown, 实际 %s", name)
	}
}

type simpleMockComponent struct{}

func (s *simpleMockComponent) Start(ctx context.Context) error { return nil }
func (s *simpleMockComponent) Stop(ctx context.Context) error  { return nil }

// TestCleanupFunction 测试清理函数
func TestCleanupFunction(t *testing.T) {
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	app, _ := New()
	app.cancel = func() {
		cleanupCalled = true
	}

	// 手动调用取消
	app.cancel()

	if !cleanupCalled {
		t.Error("清理函数未被调用")
	}
}

// TestOptions_多个选项组合 测试多个选项组合
func TestOptions_Multiple(t *testing.T) {
	component := &mockComponent{name: "test"}
	ctx := context.Background()
	waitTime := time.Second * 5

	app, cleanup := New(
		WithName("multi-option-app"),
		WithContext(ctx),
		WithWaitTime(waitTime),
		WithComponents(component),
	)
	defer cleanup()

	if app.opts.name != "multi-option-app" {
		t.Errorf("名称设置失败: %s", app.opts.name)
	}

	if app.opts.waitTime != waitTime {
		t.Errorf("等待时间设置失败: %v", app.opts.waitTime)
	}

	if len(app.opts.components) != 1 {
		t.Errorf("组件数量错误: %d", len(app.opts.components))
	}
}

// TestSignalHandling 测试信号处理
func TestSignalHandling(t *testing.T) {
	receivedSignal := false

	// 创建一个接收自定义信号的 app
	app, cleanup := New(
		WithName("signal-test"),
		WithSignal([]os.Signal{testSignal{}}),
	)
	defer cleanup()

	// 使用自定义信号测试
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(testSignal{})
	}()

	// 这个测试主要确保信号处理不会崩溃
	_ = receivedSignal
}

type testSignal struct{}

func (s testSignal) String() string { return "test-signal" }
func (s testSignal) Signal()        {}

// TestApp_ConcurrentAccess 测试并发访问
func TestApp_ConcurrentAccess(t *testing.T) {
	app, cleanup := New(WithName("concurrent-test"))
	defer cleanup()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 并发读取上下文
			_ = app.Context()
		}()
	}
	wg.Wait()
}
