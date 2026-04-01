package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jiujuan/go-ants/pkg/log"
)

// Task 任务接口
type Task interface {
	// Do 执行任务
	Do() error
	// Name 获取任务名称
	Name() string
}

// TaskFunc 任务函数类型
type TaskFunc func() error

// NamedTask 带名称的任务
type NamedTask struct {
	name string
	fn   TaskFunc
}

// NewNamedTask 创建带名称的任务
func NewNamedTask(name string, fn TaskFunc) *NamedTask {
	return &NamedTask{name: name, fn: fn}
}

// Do 执行任务
func (t *NamedTask) Do() error {
	return t.fn()
}

// Name 获取任务名称
func (t *NamedTask) Name() string {
	return t.name
}

// Pool 工作池
type Pool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	started   int32
	stopped   int32
	opts      *Options
	stats     *Stats
}

// Stats 工作池统计
type Stats struct {
	SubmittedTasks int64
	CompletedTasks int64
	FailedTasks    int64
	RunningWorkers int32
	mu             sync.RWMutex
}

// New 创建工作池
func New(workers int, opts ...Option) *Pool {
	options := &Options{
		queueSize:    1000,
		queueTimeout: time.Second * 5,
		panicHandler: defaultPanicHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		workers:   workers,
		taskQueue: make(chan Task, options.queueSize),
		ctx:       ctx,
		cancel:    cancel,
		opts:      options,
		stats:     &Stats{},
	}

	return pool
}

// Start 启动工作池
func (p *Pool) Start() {
	if !atomic.CompareAndSwapInt32(&p.started, 0, 1) {
		return
	}

	log.Info("worker pool starting",
		log.Int("workers", p.workers),
		log.Int("queueSize", p.opts.queueSize))

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	atomic.AddInt32(&p.stats.RunningWorkers, int32(p.workers))
}

// Stop 停止工作池
func (p *Pool) Stop() {
	if !atomic.CompareAndSwapInt32(&p.stopped, 0, 1) {
		return
	}

	log.Info("worker pool stopping")

	p.cancel()
	p.wg.Wait()
	close(p.taskQueue)

	log.Info("worker pool stopped")
}

// Submit 提交任务
func (p *Pool) Submit(task Task) error {
	if atomic.LoadInt32(&p.stopped) == 1 {
		return fmt.Errorf("pool is stopped")
	}

	select {
	case p.taskQueue <- task:
		atomic.AddInt64(&p.stats.SubmittedTasks, 1)
		return nil
	case <-time.After(p.opts.queueTimeout):
		return fmt.Errorf("task queue is full")
	}
}

// SubmitFunc 提交任务函数
func (p *Pool) SubmitFunc(name string, fn TaskFunc) error {
	return p.Submit(NewNamedTask(name, fn))
}

// SubmitWithContext 提交带上下文的任务
func (p *Pool) SubmitWithContext(ctx context.Context, task Task) error {
	if atomic.LoadInt32(&p.stopped) == 1 {
		return fmt.Errorf("pool is stopped")
	}

	select {
	case p.taskQueue <- task:
		atomic.AddInt64(&p.stats.SubmittedTasks, 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.opts.queueTimeout):
		return fmt.Errorf("task queue is full")
	}
}

// worker 工作协程
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	log.Debug("worker started", log.Int("workerID", id))

	for {
		select {
		case <-p.ctx.Done():
			log.Debug("worker stopped", log.Int("workerID", id))
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			p.executeTask(task)
		}
	}
}

// executeTask 执行任务
func (p *Pool) executeTask(task Task) {
	startTime := time.Now()

	log.Debug("task started",
		log.String("taskName", task.Name()))

	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
				p.opts.panicHandler(r)
			}
		}()
		return task.Do()
	}()

	duration := time.Since(startTime)

	if err != nil {
		atomic.AddInt64(&p.stats.FailedTasks, 1)
		log.Error("task failed",
			log.String("taskName", task.Name()),
			log.Error(err),
			log.Duration("duration", duration))
	} else {
		atomic.AddInt64(&p.stats.CompletedTasks, 1)
		log.Debug("task completed",
			log.String("taskName", task.Name()),
			log.Duration("duration", duration))
	}
}

// Stats 获取统计信息
func (p *Pool) Stats() Stats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	return Stats{
		SubmittedTasks: atomic.LoadInt64(&p.stats.SubmittedTasks),
		CompletedTasks: atomic.LoadInt64(&p.stats.CompletedTasks),
		FailedTasks:    atomic.LoadInt64(&p.stats.FailedTasks),
		RunningWorkers: atomic.LoadInt32(&p.stats.RunningWorkers),
	}
}

// RunningWorkers 获取运行中的 worker 数量
func (p *Pool) RunningWorkers() int {
	return int(atomic.LoadInt32(&p.stats.RunningWorkers))
}

// QueueLength 获取队列长度
func (p *Pool) QueueLength() int {
	return len(p.taskQueue)
}

// ===== 选项配置 =====

// Option 是工作池选项函数
type Option func(*Options)

type Options struct {
	queueSize    int
	queueTimeout time.Duration
	panicHandler func(interface{})
}

// WithQueueSize 设置队列大小
func WithQueueSize(size int) Option {
	return func(o *Options) {
		o.queueSize = size
	}
}

// WithQueueTimeout 设置队列超时
func WithQueueTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.queueTimeout = timeout
	}
}

// WithPanicHandler 设置 panic 处理器
func WithPanicHandler(handler func(interface{})) Option {
	return func(o *Options) {
		o.panicHandler = handler
	}
}

// defaultPanicHandler 默认 panic 处理器
func defaultPanicHandler(r interface{}) {
	log.Error("worker pool panic",
		log.Any("panic", r))
}

// ===== 任务组 =====

// Group 任务组
type Group struct {
	pool *Pool
	wg   sync.WaitGroup
}

// NewGroup 创建任务组
func NewGroup(pool *Pool) *Group {
	return &Group{pool: pool}
}

// Add 添加任务
func (g *Group) Add(task Task) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.pool.Submit(task)
	}()
}

// AddFunc 添加任务函数
func (g *Group) AddFunc(name string, fn TaskFunc) {
	g.Add(NewNamedTask(name, fn))
}

// Wait 等待完成
func (g *Group) Wait() {
	g.wg.Wait()
}

// ===== 带结果的任务 =====

// ResultTask 带结果的任务
type ResultTask struct {
	name   string
	fn     func() (interface{}, error)
	result interface{}
	err    error
	once   sync.Once
}

// NewResultTask 创建带结果的任务
func NewResultTask(name string, fn func() (interface{}, error)) *ResultTask {
	return &ResultTask{name: name, fn: fn}
}

// Do 执行任务
func (t *ResultTask) Do() error {
	t.once.Do(func() {
		t.result, t.err = t.fn()
	})
	return t.err
}

// Name 获取任务名称
func (t *ResultTask) Name() string {
	return t.name
}

// Result 获取结果
func (t *ResultTask) Result() (interface{}, error) {
	return t.result, t.err
}

// ===== 定期任务 =====

// Scheduler 任务调度器
type Scheduler struct {
	pool   *Pool
	tasks  map[string]*scheduledTask
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// scheduledTask 定期任务
type scheduledTask struct {
	interval time.Duration
	task     Task
}

// NewScheduler 创建调度器
func NewScheduler(pool *Pool) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		pool:   pool,
		tasks:  make(map[string]*scheduledTask),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Schedule 调度任务
func (s *Scheduler) Schedule(name string, interval time.Duration, task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = &scheduledTask{
		interval: interval,
		task:     task,
	}

	go s.runTask(name, interval, task)
}

// ScheduleFunc 调度任务函数
func (s *Scheduler) ScheduleFunc(name string, interval time.Duration, fn TaskFunc) {
	s.Schedule(name, interval, NewNamedTask(name, fn))
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
}

func (s *Scheduler) runTask(name string, interval time.Duration, task Task) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.pool.Submit(task)
		}
	}
}

// ===== 限流器 =====

// Limiter 限流器
type Limiter struct {
	rate     int // 每秒允许的请求数
	burst    int // 突发容量
	tokens   int
	lastTime time.Time
	mu       sync.Mutex
}

// NewLimiter 创建限流器
func NewLimiter(rate int, burst int) *Limiter {
	return &Limiter{
		rate:     rate,
		burst:    burst,
		tokens:   burst,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime)
	l.tokens = l.tokens + int(elapsed.Seconds()*float64(l.rate))
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.lastTime = now

	if l.tokens > 0 {
		l.tokens--
		return true
	}

	return false
}

// Wait 等待直到允许
func (l *Limiter) Wait() {
	for !l.Allow() {
		time.Sleep(time.Millisecond * 10)
	}
}

// WaitContext 等待直到允许或上下文取消
func (l *Limiter) WaitContext(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}
