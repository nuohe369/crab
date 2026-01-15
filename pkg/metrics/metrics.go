package metrics

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Config metrics configuration
type Config struct {
	Enabled bool   `toml:"enabled"`
	Path    string `toml:"path"` // Metrics expose path, default /metrics
}

var (
	enabled     bool
	path        string
	registry    *prometheus.Registry
	initialized bool

	// Metrics cache (avoid duplicate registration)
	counters   = make(map[string]*prometheus.CounterVec)
	gauges     = make(map[string]*prometheus.GaugeVec)
	histograms = make(map[string]*prometheus.HistogramVec)
	mu         sync.RWMutex
)

// Init initializes the metrics system
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

	// Create custom registry
	registry = prometheus.NewRegistry()

	// Register Go runtime metrics
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Register HTTP request metrics
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(httpRequestsInFlight)

	log.Printf("metrics: enabled, path: %s", path)
}

// Enabled checks if metrics is enabled
func Enabled() bool {
	return enabled
}

// Path returns the metrics path
func Path() string {
	return path
}

// Registry returns the prometheus registry
func Registry() *prometheus.Registry {
	return registry
}

// ============ HTTP request metrics (built-in) ============

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
func RecordHTTPRequest(method, path, status string, duration float64) {
	if !enabled {
		return
	}
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// IncHTTPInFlight increments the in-flight request count
func IncHTTPInFlight() {
	if enabled {
		httpRequestsInFlight.Inc()
	}
}

// DecHTTPInFlight decrements the in-flight request count
func DecHTTPInFlight() {
	if enabled {
		httpRequestsInFlight.Dec()
	}
}

// ============ Custom metrics convenience methods ============

// Counter gets or creates a Counter metric
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

	// 双重检查
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

// ============ Simplified business metrics methods ============

// Inc increments a counter (no labels)
func Inc(name, help string) {
	if c := Counter(name, help); c != nil {
		c.WithLabelValues().Inc()
	}
}

// IncWithLabels increments a counter (with labels)
func IncWithLabels(name, help string, labelNames []string, labelValues ...string) {
	if c := Counter(name, help, labelNames...); c != nil {
		c.WithLabelValues(labelValues...).Inc()
	}
}

// Set sets a Gauge value (no labels)
func Set(name, help string, value float64) {
	if g := Gauge(name, help); g != nil {
		g.WithLabelValues().Set(value)
	}
}

// Observe records a Histogram observation (no labels)
func Observe(name, help string, value float64) {
	if h := Histogram(name, help, nil); h != nil {
		h.WithLabelValues().Observe(value)
	}
}
