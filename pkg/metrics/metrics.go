// Package metrics provides Prometheus metrics collection and exposure
// Package metrics 提供 Prometheus 指标收集和暴露
package metrics

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Config represents metrics configuration
// Config 表示指标配置
type Config struct {
	Enabled bool   `toml:"enabled"` // Enable metrics | 启用指标
	Path    string `toml:"path"`    // Metrics expose path, default /metrics | 指标暴露路径，默认 /metrics
}

var (
	enabled     bool                 // Whether metrics is enabled | 指标是否启用
	path        string               // Metrics path | 指标路径
	registry    *prometheus.Registry // Prometheus registry | Prometheus 注册表
	initialized bool                 // Whether initialized | 是否已初始化

	// Metrics cache (avoid duplicate registration) | 指标缓存（避免重复注册）
	counters   = make(map[string]*prometheus.CounterVec)
	gauges     = make(map[string]*prometheus.GaugeVec)
	histograms = make(map[string]*prometheus.HistogramVec)
	mu         sync.RWMutex // Mutex for concurrent access | 并发访问互斥锁
)

// Init initializes the metrics system
// Init 初始化指标系统
func Init(cfg Config) {
	if initialized {
		return
	}
	initialized = true

	enabled = cfg.Enabled
	path = cfg.Path
	if path == "" {
		path = "/metrics"
	}

	if !enabled {
		log.Println("metrics: not enabled")
		return
	}

	// Create custom registry | 创建自定义注册表
	registry = prometheus.NewRegistry()

	// Register Go runtime metrics | 注册 Go 运行时指标
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Register HTTP request metrics | 注册 HTTP 请求指标
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(httpRequestsInFlight)

	log.Printf("metrics: enabled, path: %s", path)
}

// Enabled checks if metrics is enabled
// Enabled 检查指标是否启用
func Enabled() bool {
	return enabled
}

// Path returns the metrics path
// Path 返回指标路径
func Path() string {
	return path
}

// Registry returns the prometheus registry
// Registry 返回 prometheus 注册表
func Registry() *prometheus.Registry {
	return registry
}

// ============ HTTP request metrics (built-in) | HTTP 请求指标（内置）============

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
)

// RecordHTTPRequest records HTTP request metrics
// RecordHTTPRequest 记录 HTTP 请求指标
func RecordHTTPRequest(method, path, status string, duration float64) {
	if !enabled {
		return
	}
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// IncHTTPInFlight increments the in-flight request count
// IncHTTPInFlight 增加正在处理的请求计数
func IncHTTPInFlight() {
	if enabled {
		httpRequestsInFlight.Inc()
	}
}

// DecHTTPInFlight decrements the in-flight request count
// DecHTTPInFlight 减少正在处理的请求计数
func DecHTTPInFlight() {
	if enabled {
		httpRequestsInFlight.Dec()
	}
}

// ============ Custom metrics convenience methods | 自定义指标便捷方法 ============

// Counter gets or creates a Counter metric
// Counter 获取或创建 Counter 指标
func Counter(name, help string, labels ...string) *prometheus.CounterVec {
	if !enabled {
		return nil
	}

	mu.RLock()
	if c, ok := counters[name]; ok {
		mu.RUnlock()
		return c
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Double check | 双重检查
	if c, ok := counters[name]; ok {
		return c
	}

	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	registry.MustRegister(c)
	counters[name] = c
	return c
}

// Gauge gets or creates a Gauge metric
// Gauge 获取或创建 Gauge 指标
func Gauge(name, help string, labels ...string) *prometheus.GaugeVec {
	if !enabled {
		return nil
	}

	mu.RLock()
	if g, ok := gauges[name]; ok {
		mu.RUnlock()
		return g
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if g, ok := gauges[name]; ok {
		return g
	}

	g := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	registry.MustRegister(g)
	gauges[name] = g
	return g
}

// Histogram gets or creates a Histogram metric
// Histogram 获取或创建 Histogram 指标
func Histogram(name, help string, buckets []float64, labels ...string) *prometheus.HistogramVec {
	if !enabled {
		return nil
	}

	mu.RLock()
	if h, ok := histograms[name]; ok {
		mu.RUnlock()
		return h
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if h, ok := histograms[name]; ok {
		return h
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	h := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	registry.MustRegister(h)
	histograms[name] = h
	return h
}

// ============ Simplified business metrics methods | 简化的业务指标方法 ============

// Inc increments a counter (no labels)
// Inc 增加计数器（无标签）
func Inc(name, help string) {
	if c := Counter(name, help); c != nil {
		c.WithLabelValues().Inc()
	}
}

// IncWithLabels increments a counter (with labels)
// IncWithLabels 增加计数器（带标签）
func IncWithLabels(name, help string, labelNames []string, labelValues ...string) {
	if c := Counter(name, help, labelNames...); c != nil {
		c.WithLabelValues(labelValues...).Inc()
	}
}

// Set sets a Gauge value (no labels)
// Set 设置 Gauge 值（无标签）
func Set(name, help string, value float64) {
	if g := Gauge(name, help); g != nil {
		g.WithLabelValues().Set(value)
	}
}

// Observe records a Histogram observation (no labels)
// Observe 记录 Histogram 观察值（无标签）
func Observe(name, help string, value float64) {
	if h := Histogram(name, help, nil); h != nil {
		h.WithLabelValues().Observe(value)
	}
}
