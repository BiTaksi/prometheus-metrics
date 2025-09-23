package metrics

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	Namespace string
	Subsystem string
}

// DefaultConfig returns a default metrics configuration
func DefaultConfig() *MetricsConfig {
	return &MetricsConfig{
		Namespace: "",
		Subsystem: "",
	}
}

// Metrics holds all prometheus metrics
type Metrics struct {
	// HTTP request metrics
	ProcessedOpsTotal *prometheus.CounterVec
	FailedOpsTotal    *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec

	// Service health status metric
	ServiceUp prometheus.Gauge

	// System resource metrics
	MemoryUsageBytes prometheus.Gauge
	GoroutinesCount  prometheus.Gauge
	GCDuration       prometheus.Histogram

	// Business metrics - extensible counters and gauges
	BusinessCounters   map[string]*prometheus.CounterVec
	BusinessGauges     map[string]*prometheus.GaugeVec
	BusinessHistograms map[string]*prometheus.HistogramVec
}

// NewMetrics creates a new Metrics instance with the given configuration
func NewMetrics(config *MetricsConfig) *Metrics {
	if config == nil {
		config = DefaultConfig()
	}

	return &Metrics{
		ProcessedOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "The total number of processed operations",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		FailedOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_failed_total",
				Help:      "The total number of failed operations",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),

		ServiceUp: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "service_up",
				Help:      "Whether the service is up (1) or down (0)",
			},
		),

		MemoryUsageBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),

		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "goroutines_count",
				Help:      "Number of active goroutines",
			},
		),

		GCDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "gc_duration_seconds",
				Help:      "Garbage collection duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
		),

		// Initialize business metrics maps
		BusinessCounters:   make(map[string]*prometheus.CounterVec),
		BusinessGauges:     make(map[string]*prometheus.GaugeVec),
		BusinessHistograms: make(map[string]*prometheus.HistogramVec),
	}
}

// HTTPMiddleware returns a standard HTTP middleware function for Prometheus metrics
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	// Mark service as up when it starts
	m.ServiceUp.Set(1)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response recorder to capture status code
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(recorder, r)

		// Collect metric information
		method := r.Method
		endpoint := r.URL.Path
		statusCode := strconv.Itoa(recorder.statusCode)
		duration := time.Since(start).Seconds()

		// Update metrics
		m.ProcessedOpsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		m.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)

		// Increment failed operations metric on error
		if recorder.statusCode >= 400 {
			errorType := "client_error"
			if recorder.statusCode >= 500 {
				errorType = "server_error"
			}
			m.FailedOpsTotal.WithLabelValues(method, endpoint, errorType).Inc()
		}
	})
}

// responseRecorder is a wrapper around http.ResponseWriter to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// SetServiceDown marks service as down when shutting down
func (m *Metrics) SetServiceDown() {
	m.ServiceUp.Set(0)
}

// UpdateSystemMetrics updates system resource metrics
func (m *Metrics) UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage
	m.MemoryUsageBytes.Set(float64(memStats.Alloc))

	// Update goroutine count
	m.GoroutinesCount.Set(float64(runtime.NumGoroutine()))
}

// RecordGCDuration records garbage collection duration
func (m *Metrics) RecordGCDuration(duration time.Duration) {
	m.GCDuration.Observe(duration.Seconds())
}

// AddBusinessCounter creates a new business counter metric
func (m *Metrics) AddBusinessCounter(name, help string, labels []string) *prometheus.CounterVec {
	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.BusinessCounters[name] = counter
	return counter
}

// AddBusinessGauge creates a new business gauge metric
func (m *Metrics) AddBusinessGauge(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.BusinessGauges[name] = gauge
	return gauge
}

// AddBusinessHistogram creates a new business histogram metric
func (m *Metrics) AddBusinessHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	m.BusinessHistograms[name] = histogram
	return histogram
}

// GetBusinessCounter returns a business counter by name
func (m *Metrics) GetBusinessCounter(name string) (*prometheus.CounterVec, bool) {
	counter, exists := m.BusinessCounters[name]
	return counter, exists
}

// GetBusinessGauge returns a business gauge by name
func (m *Metrics) GetBusinessGauge(name string) (*prometheus.GaugeVec, bool) {
	gauge, exists := m.BusinessGauges[name]
	return gauge, exists
}

// GetBusinessHistogram returns a business histogram by name
func (m *Metrics) GetBusinessHistogram(name string) (*prometheus.HistogramVec, bool) {
	histogram, exists := m.BusinessHistograms[name]
	return histogram, exists
}

// IncrementBusinessCounter increments a business counter
func (m *Metrics) IncrementBusinessCounter(name string, labelValues ...string) {
	if counter, exists := m.BusinessCounters[name]; exists {
		counter.WithLabelValues(labelValues...).Inc()
	}
}

// SetBusinessGauge sets a business gauge value
func (m *Metrics) SetBusinessGauge(name string, value float64, labelValues ...string) {
	if gauge, exists := m.BusinessGauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Set(value)
	}
}

// ObserveBusinessHistogram observes a value in business histogram
func (m *Metrics) ObserveBusinessHistogram(name string, value float64, labelValues ...string) {
	if histogram, exists := m.BusinessHistograms[name]; exists {
		histogram.WithLabelValues(labelValues...).Observe(value)
	}
}
