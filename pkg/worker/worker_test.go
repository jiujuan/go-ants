package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNew_创建工作池 测试创建工作池
func TestNew_创建工作池(t *testing.T) {
	pool := New(10)

	if pool == nil {
		t.Error("工作池创建失败")
	}

	if pool.workers != 10 {
		t.Errorf("工作协程数量期望 10, 得到: %d", pool.workers)
	}
}

// TestNew_默认选项 测试默认选项创建工作池
func TestNew_默认选项(t *testing.T) {
	pool := New(5)

	if pool.workers != 5 {
		t.Error("工作协程数量设置失败")
	}

	if pool.taskQueue == nil {
		t.Error("任务队列未创建")
	}

	if pool.ctx == nil {
		t.Error("上下文未创建")
	}
}

// TestWithQueueSize_选项函数 测试队列大小选项函数
func TestWithQueueSize_选项函数(t *testing.T) {
	pool := New(5, WithQueueSize(1000))

	// 检查队列容量
	if cap(pool.taskQueue) != 1000 {
		t.Errorf("队列大小期望 1000, 得到: %d", cap(pool.taskQueue))
	}
}

// TestWithQueueTimeout_选项函数 测试队列超时选项函数
func TestWithQueueTimeout_选项函数(t *testing.T) {
	timeout := time.Second * 10
	pool := New(5, WithQueueTimeout(timeout))

	if pool.opts.queueTimeout != timeout {
		t.Error("队列超时设置失败")
	}
}

// TestWithPanicHandler_选项函数 测试 panic 处理器选项函数
func TestWithPanicHandler_选项函数(t *testing.T) {
	panicCalled := false
	handler := func(interface{}) {
		panicCalled = true
	}

	pool := New(5, WithPanicHandler(handler))

	_ = panicCalled // 使用变量避免警告
	_ = pool        // 使用变量避免警告
}

// TestPool_启动和停止 测试工作池启动和停止
func TestPool_启动和停止(t *testing.T) {
	pool := New(2)

	// 启动工作池
	pool.Start()

	// 验证已启动
	if atomic.LoadInt32(&pool.started) != 1 {
		t.Error("工作池未启动")
	}

	// 停止工作池
	pool.Stop()

	// 验证已停止
	if atomic.LoadInt32(&pool.stopped) != 1 {
		t.Error("工作池未停止")
	}
}

// TestPool_重复启动 测试重复启动
func TestPool_重复启动(t *testing.T) {
	pool := New(2)

	pool.Start()
	pool.Start() // 重复启动

	// 应该只启动一次
	if atomic.LoadInt32(&pool.started) != 1 {
		t.Error("重复启动应该无效")
	}
}

// TestPool_重复停止 测试重复停止
func TestPool_重复停止(t *testing.T) {
	pool := New(2)

	pool.Start()
	pool.Stop()
	pool.Stop() // 重复停止

	// 应该只停止一次
	if atomic.LoadInt32(&pool.stopped) != 1 {
		t.Error("重复停止应该无效")
	}
}

// TestSubmit_提交任务 测试提交任务
func TestSubmit_提交任务(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	task := &mockTask{name: "test-task", executeErr: nil}

	err := pool.Submit(task)

	if err != nil {
		t.Errorf("任务提交失败: %v", err)
	}

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	if !task.executed {
		t.Error("任务未执行")
	}
}

// TestSubmit_队列满 测试队列满的情况
func TestSubmit_队列满(t *testing.T) {
	// 创建一个队列很小的工作池
	pool := New(1, WithQueueSize(1))
	pool.Start()
	defer pool.Stop()

	// 提交一个会长时间运行的任务
	longTask := &mockTask{
		name:  "long-task",
		delay: time.Hour, // 长时间延迟
	}
	pool.Submit(longTask)

	// 尝试提交第二个任务，队列应该满了
	err := pool.Submit(&mockTask{name: "should-fail"})

	if err == nil {
		t.Error("应该返回队列满的错误")
	}
}

// TestSubmitFunc_提交任务函数 测试提交任务函数
func TestSubmitFunc_提交任务函数(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	executed := false
	err := pool.SubmitFunc("func-task", func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("任务提交失败: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("任务函数未执行")
	}
}

// TestSubmit_已停止的池 测试向已停止的池提交任务
func TestSubmit_已停止的池(t *testing.T) {
	pool := New(2)
	pool.Start()
	pool.Stop()

	task := &mockTask{name: "test-task"}
	err := pool.Submit(task)

	if err == nil {
		t.Error("应该返回池已停止的错误")
	}
}

// TestStats_统计信息 测试统计信息
func TestStats_统计信息(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	// 提交一些任务
	for i := 0; i < 5; i++ {
		pool.Submit(&mockTask{name: "task", executeErr: nil})
	}

	time.Sleep(200 * time.Millisecond)

	stats := pool.Stats()

	if stats.SubmittedTasks != 5 {
		t.Errorf("提交任务数期望 5, 得到: %d", stats.SubmittedTasks)
	}

	if stats.CompletedTasks < 5 {
		t.Errorf("完成任务数应该 >= 5, 得到: %d", stats.CompletedTasks)
	}
}

// TestRunningWorkers_运行工作协程数 测试运行中的工作协程数
func TestRunningWorkers_运行工作协程数(t *testing.T) {
	pool := New(5)
	pool.Start()

	count := pool.RunningWorkers()

	if count != 5 {
		t.Errorf("运行工作协程数期望 5, 得到: %d", count)
	}

	pool.Stop()

	count = pool.RunningWorkers()
	if count != 0 {
		t.Errorf("停止后期望 0, 得到: %d", count)
	}
}

// TestQueueLength_队列长度 测试队列长度
func TestQueueLength_队列长度(t *testing.T) {
	pool := New(1, WithQueueSize(10))
	pool.Start()
	defer pool.Stop()

	// 提交任务填满队列
	pool.Submit(&mockTask{name: "long-task", delay: time.Hour})

	length := pool.QueueLength()

	if length != 1 {
		t.Errorf("队列长度期望 1, 得到: %d", length)
	}
}

// TestNamedTask_带名称的任务 测试带名称的任务
func TestNamedTask_带名称的任务(t *testing.T) {
	task := NewNamedTask("my-task", func() error {
		return nil
	})

	if task.Name() != "my-task" {
		t.Errorf("任务名称期望 my-task, 得到: %s", task.Name())
	}

	err := task.Do()
	if err != nil {
		t.Errorf("任务执行失败: %v", err)
	}
}

// TestNamedTask_执行错误 测试任务执行错误
func TestNamedTask_执行错误(t *testing.T) {
	taskErr := errors.New("task error")
	task := NewNamedTask("error-task", func() error {
		return taskErr
	})

	err := task.Do()

	if err != taskErr {
		t.Errorf("期望任务错误, 得到: %v", err)
	}
}

// TestGroup_任务组 测试任务组
func TestGroup_任务组(t *testing.T) {
	pool := New(5)
	pool.Start()
	defer pool.Stop()

	group := NewGroup(pool)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		group.AddFunc("task", func() error {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}

	// 由于 Add 是异步的，这里只验证不崩溃
	time.Sleep(100 * time.Millisecond)
}

// TestResultTask_带结果的任务 测试带结果的任务
func TestResultTask_带结果的任务(t *testing.T) {
	task := NewResultTask("result-task", func() (interface{}, error) {
		return "result", nil
	})

	err := task.Do()
	if err != nil {
		t.Errorf("任务执行失败: %v", err)
	}

	result, err := task.Result()
	if err != nil {
		t.Errorf("获取结果失败: %v", err)
	}

	if result != "result" {
		t.Errorf("结果期望 'result', 得到: %v", result)
	}
}

// TestResultTask_多次获取结果 测试多次获取结果
func TestResultTask_多次获取结果(t *testing.T) {
	task := NewResultTask("once-task", func() (interface{}, error) {
		return "once", nil
	})

	// 第一次获取
	task.Do()
	result1, _ := task.Result()

	// 第二次获取 - 应该返回相同结果
	result2, _ := task.Result()

	if result1 != result2 {
		t.Error("多次获取结果应该返回相同值")
	}
}

// TestLimiter_限流器 测试限流器
func TestLimiter_限流器(t *testing.T) {
	limiter := NewLimiter(10, 5)

	// 突发容量应该是 5
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Errorf("第 %d 次允许失败", i)
		}
	}

	// 第 6 次应该被限流
	if limiter.Allow() {
		t.Error("第 6 次应该被限流")
	}
}

// TestLimiter_Wait  测试等待限流
func TestLimiter_Wait(t *testing.T) {
	limiter := NewLimiter(1000, 1) // 高速率，小突发

	// 消耗突发容量
	limiter.Allow()

	// 等待后应该可以继续
	limiter.Wait()

	if !limiter.Allow() {
		t.Error("等待后应该允许")
	}
}

// TestLimiter_WaitContext 测试带上下文的等待限流
func TestLimiter_WaitContext(t *testing.T) {
	limiter := NewLimiter(1000, 1)

	// 消耗突发容量
	limiter.Allow()

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 应该立即返回上下文错误
	err := limiter.WaitContext(ctx)

	if err == nil {
		t.Error("应该返回上下文错误")
	}
}

// TestOptions_多个选项组合 测试多个选项组合
func TestOptions_多个选项组合(t *testing.T) {
	queueSize := 500
	timeout := time.Second * 10

	pool := New(10,
		WithQueueSize(queueSize),
		WithQueueTimeout(timeout),
	)

	if pool.workers != 10 {
		t.Error("工作协程数设置失败")
	}

	if cap(pool.taskQueue) != queueSize {
		t.Errorf("队列大小设置失败: %d", cap(pool.taskQueue))
	}

	if pool.opts.queueTimeout != timeout {
		t.Error("队列超时设置失败")
	}
}

// TestTaskExecution_任务执行 测试任务执行
func TestTaskExecution_任务执行(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	executedTasks := make(chan string, 10)

	// 提交多个任务
	for i := 0; i < 5; i++ {
		taskName := "task"
		pool.SubmitFunc(taskName, func() error {
			executedTasks <- taskName
			return nil
		})
	}

	// 等待任务完成
	timeout := time.After(2 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case <-executedTasks:
		case <-timeout:
			t.Error("任务执行超时")
			return
		}
	}
}

// TestTaskError_任务错误 测试任务错误处理
func TestTaskError_任务错误(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	errTask := NewNamedTask("error-task", func() error {
		return errors.New("intentional error")
	})

	pool.Submit(errTask)

	time.Sleep(100 * time.Millisecond)

	stats := pool.Stats()
	if stats.FailedTasks < 1 {
		t.Error("失败任务数应该 >= 1")
	}
}

// mockTask 是一个模拟任务
type mockTask struct {
	name       string
	executed   bool
	executeErr error
	delay      time.Duration
	mu         sync.Mutex
}

func (m *mockTask) Do() error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.Lock()
	m.executed = true
	m.mu.Unlock()

	return m.executeErr
}

func (m *mockTask) Name() string {
	return m.name
}

// TestScheduler_调度器 测试调度器
func TestScheduler_调度器(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	scheduler := NewScheduler(pool)
	defer scheduler.Stop()

	taskExecuted := false
	scheduler.ScheduleFunc("periodic-task", time.Second, func() error {
		taskExecuted = true
		return nil
	})

	// 等待至少执行一次
	time.Sleep(1500 * time.Millisecond)

	if !taskExecuted {
		t.Error("定期任务未执行")
	}
}

// TestScheduler_多个任务 测试调度器多个任务
func TestScheduler_多个任务(t *testing.T) {
	pool := New(2)
	pool.Start()
	defer pool.Stop()

	scheduler := NewScheduler(pool)
	defer scheduler.Stop()

	scheduler.ScheduleFunc("task1", time.Second, func() error { return nil })
	scheduler.ScheduleFunc("task2", time.Second*2, func() error { return nil })

	// 等待任务调度
	time.Sleep(100 * time.Millisecond)

	// 不崩溃即通过
}
