package metric

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/jiujuan/go-ants/pkg/log"
)

// Metrics 全局指标实例
var (
	// RequestCounter 请求计数器
	RequestCounter *prometheus.CounterVec
	// RequestDuration 请求延迟直方图
	RequestDuration *prometheus.HistogramVec
	// RequestInFlight 当前请求数
	RequestInFlight *prometheus.GaugeVec

	// DBQueryCounter 数据库查询计数器
	DBQueryCounter *prometheus.CounterVec
	// DBQueryDuration 数据库查询延迟直方图
	DBQueryDuration *prometheus.HistogramVec

	// RedisCounter Redis 操作计数器
	RedisCounter *prometheus.CounterVec
	// RedisDuration Redis 操作延迟直方图
	RedisDuration *prometheus.HistogramVec

	// WorkerPoolSize 工作池大小
	WorkerPoolSize *prometheus.GaugeVec
	// WorkerPoolCapacity 工作池容量
	WorkerPoolCapacity *prometheus.GaugeVec
	// WorkerPoolTaskDuration 工作池任务延迟
	WorkerPoolTaskDuration *prometheus.HistogramVec

	// CustomCounter 自定义计数器
	CustomCounter *prometheus.CounterVec
	// CustomGauge 自定义仪表
	CustomGauge *prometheus.GaugeVec
	// CustomHistogram 自定义直方图
	CustomHistogram *prometheus.HistogramVec
)

var initOnce sync.Once

// InitMetrics 初始化指标
func InitMetrics(serviceName string) {
	initOnce.Do(func() {
		RequestCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_http_requests_total", serviceName),
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)

		RequestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_http_request_duration_seconds", serviceName),
				Help:    "HTTP request latency in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"method", "path"},
		)

		RequestInFlight = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_http_requests_in_flight", serviceName),
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "path"},
		)

		DBQueryCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_db_queries_total", serviceName),
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		)

		DBQueryDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_db_query_duration_seconds", serviceName),
				Help:    "Database query latency in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		)

		RedisCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_redis_operations_total", serviceName),
				Help: "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		)

		RedisDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_redis_operation_duration_seconds", serviceName),
				Help:    "Redis operation latency in seconds",
				Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
			},
			[]string{"operation"},
		)

		WorkerPoolSize = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_worker_pool_size", serviceName),
				Help: "Current number of workers in the pool",
			},
			[]string{"pool"},
		)

		WorkerPoolCapacity = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_worker_pool_capacity", serviceName),
				Help: "Maximum number of workers in the pool",
			},
			[]string{"pool"},
		)

		WorkerPoolTaskDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_worker_task_duration_seconds", serviceName),
				Help:    "Worker task execution latency in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"pool", "task"},
		)

		CustomCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_custom_counter_total", serviceName),
				Help: "Custom counter metric",
			},
			[]string{"name", "labels"},
		)

		CustomGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_custom_gauge", serviceName),
				Help: "Custom gauge metric",
			},
			[]string{"name", "labels"},
		)

		CustomHistogram = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_custom_histogram", serviceName),
				Help:    "Custom histogram metric",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"name", "labels"},
		)

		// 启动 Go 运行时指标收集
		go collectGoMetrics(serviceName)
		go collectProcessMetrics(serviceName)

		log.Info("metrics initialized", log.String("service", serviceName))
	})
}

// ===== HTTP 中间件指标 =====

// HTTPMetrics HTTP 指标中间件
type HTTPMetrics struct {
	Method string
	Path   string
}

// NewHTTPMetrics 创建 HTTP 指标中间件
func NewHTTPMetrics(method, path string) *HTTPMetrics {
	return &HTTPMetrics{
		Method: method,
		Path:   path,
	}
}

// Observe 记录观察值
func (m *HTTPMetrics) Observe(duration time.Duration) {
	RequestDuration.WithLabelValues(m.Method, m.Path).Observe(duration.Seconds())
}

// Inc 增加计数器
func (m *HTTPMetrics) Inc(status string) {
	RequestCounter.WithLabelValues(m.Method, m.Path, status).Inc()
}

// SetInFlight 设置当前请求数
func (m *HTTPMetrics) SetInFlight(i float64) {
	RequestInFlight.WithLabelValues(m.Method, m.Path).Set(i)
}

// ===== 数据库指标 =====

// DBMetrics 数据库指标
type DBMetrics struct {
	Operation string
	Table     string
}

// NewDBMetrics 创建数据库指标
func NewDBMetrics(operation, table string) *DBMetrics {
	return &DBMetrics{
		Operation: operation,
		Table:     table,
	}
}

// Observe 记录查询延迟
func (m *DBMetrics) Observe(duration time.Duration) {
	DBQueryDuration.WithLabelValues(m.Operation, m.Table).Observe(duration.Seconds())
}

// Inc 增加计数器
func (m *DBMetrics) Inc(status string) {
	DBQueryCounter.WithLabelValues(m.Operation, m.Table, status).Inc()
}

// ===== Redis 指标 =====

// RedisMetrics Redis 指标
type RedisMetrics struct {
	Operation string
}

// NewRedisMetrics 创建 Redis 指标
func NewRedisMetrics(operation string) *RedisMetrics {
	return &RedisMetrics{
		Operation: operation,
	}
}

// Observe 记录操作延迟
func (m *RedisMetrics) Observe(duration time.Duration) {
	RedisDuration.WithLabelValues(m.Operation).Observe(duration.Seconds())
}

// Inc 增加计数器
func (m *RedisMetrics) Inc(status string) {
	RedisCounter.WithLabelValues(m.Operation, status).Inc()
}

// ===== Go 运行时指标 =====

func collectGoMetrics(serviceName string) {
	// Go 版本信息
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_go_info", serviceName),
			Help: "Go version information",
		},
		[]string{"version"},
	)
	prometheus.MustRegister(buildInfo)

	info := runtime.Version()
	buildInfo.WithLabelValues(info).Set(1)

	// Goroutine 数量
	goroutines := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_go_goroutines", serviceName),
		Help: "Number of goroutines",
	})
	prometheus.MustRegister(goroutines)

	// GC 信息
	gcDesc := prometheus.NewDesc(
		fmt.Sprintf("%s_go_gc_duration_seconds", serviceName),
		"Go garbage collection duration in seconds",
		nil, nil,
	)
	gcLast := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_go_gc_last", serviceName),
		Help: "Go last garbage collection duration in seconds",
	})
	prometheus.MustRegister(gcLast)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		goroutines.Set(float64(runtime.NumGoroutine()))

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		gcLast.Set(float64(stats.GCLastNum) / float64(1e9))
	}
}

// ===== 进程指标 =====

func collectProcessMetrics(serviceName string) {
	// 进程信息
	procDesc := prometheus.NewDesc(
		fmt.Sprintf("%s_process_start_time_seconds", serviceName),
		"Start time of the process in seconds since epoch",
		nil, nil,
	)
	procStart := prometheus.NewGauge(procDesc)
	prometheus.MustRegister(procStart)

	procStart.Set(float64(time.Now().Unix()))

	// 打开文件描述符
	openFDs := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_process_open_fds", serviceName),
		Help: "Number of open file descriptors",
	})
	prometheus.MustRegister(openFDs)

	// 最大文件描述符
	maxFDs := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_process_max_fds", serviceName),
		Help: "Maximum number of open file descriptors",
	})
	prometheus.MustRegister(maxFDs)

	// 内存信息
	memDesc := prometheus.NewDesc(
		fmt.Sprintf("%s_process_resident_memory_bytes", serviceName),
		"Resident memory usage in bytes",
		nil, nil,
	)
	memUsage := prometheus.NewGauge(memDesc)
	prometheus.MustRegister(memUsage)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		memUsage.Set(float64(stats.Alloc))
	}
}

// ===== 自定义指标工具 =====

// IncCounter 增加计数器
func IncCounter(name, labels string) {
	CustomCounter.WithLabelValues(name, labels).Inc()
}

// SetGauge 设置仪表值
func SetGauge(name, labels string, value float64) {
	CustomGauge.WithLabelValues(name, labels).Set(value)
}

// IncGauge 增加仪表值
func IncGauge(name, labels string) {
	CustomGauge.WithLabelValues(name, labels).Inc()
}

// DecGauge 减少仪表值
func DecGauge(name, labels string) {
	CustomGauge.WithLabelValues(name, labels).Dec()
}

// ObserveHistogram 观察直方图
func ObserveHistogram(name, labels string, duration time.Duration) {
	CustomHistogram.WithLabelValues(name, labels).Observe(duration.Seconds())
}

// ===== PProf 集成 =====

// StartPProf 启动 PProf
func StartPProf(addr string) {
	go func() {
		log.Info("pprof server starting",
			log.String("addr", addr))
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Error("pprof server error",
				log.Error(err))
		}
	}()
}

// NewPProfServer 创建 PProf 服务器
func NewPProfServer(addr string) *http.Server {
	return &http.Server{
		Addr: addr,
	}
}
