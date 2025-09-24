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
	// Normalize is an optional function to transform the request into a low-cardinality endpoint label
	// If nil, it defaults to returning r.URL.Path as-is
	Normalize EndpointNormalizer
}

// DefaultConfig returns a default metrics configuration
func DefaultConfig() *MetricsConfig {
	return &MetricsConfig{
		Namespace: "",
		Subsystem: "",
	}
}

// Metrics holds all prometheus metrics
type IMetrics interface {
	IBaseMetrics
	HTTPMiddleware(next http.Handler) http.Handler
	SetServiceUp()
	SetServiceDown()
	UpdateSystemMetrics()
	RecordGCDuration(duration time.Duration)
	IncreaseProcessedOps(method, endpoint, statusCode string)
	IncreaseFailedOps(method, endpoint, errorType string)
	ObserveRequestDuration(method, endpoint, statusCode string, duration float64)
}

// EndpointNormalizer transforms a request to a normalized, low-cardinality endpoint label
type EndpointNormalizer func(r *http.Request) string

type metrics struct {
	IBaseMetrics
	// HTTP request metrics
	processedOpsTotal *prometheus.CounterVec
	failedOpsTotal    *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec

	// Service health status metric
	serviceUp prometheus.Gauge

	// System resource metrics
	memoryUsageBytes prometheus.Gauge
	goroutinesCount  prometheus.Gauge
	gcDuration       prometheus.Histogram

	// endpoint normalizer
	normalize EndpointNormalizer
}

// NewMetrics creates a new Metrics instance with the given configuration
func NewMetrics(config *MetricsConfig) IMetrics {
	if config == nil {
		config = DefaultConfig()
	}

	m := &metrics{
		IBaseMetrics: initializeBaseMetrics(),
		processedOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "The total number of processed operations",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		failedOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_failed_total",
				Help:      "The total number of failed operations",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),

		serviceUp: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "service_up",
				Help:      "Whether the service is up (1) or down (0)",
			},
		),

		memoryUsageBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),

		goroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "goroutines_count",
				Help:      "Number of active goroutines",
			},
		),

		gcDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "gc_duration_seconds",
				Help:      "Garbage collection duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
		),
	}

	// set normalizer
	if config.Normalize != nil {
		m.normalize = config.Normalize
	} else {
		m.normalize = defaultEndpointNormalize
	}

	return m
}

// HTTPMiddleware returns a standard HTTP middleware function for Prometheus metrics
func (m *metrics) HTTPMiddleware(next http.Handler) http.Handler {
	// Mark service as up when it starts
	m.serviceUp.Set(1)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response recorder to capture status code
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Ensure metrics are recorded even if the handler panics
		defer func() {
			if rec := recover(); rec != nil {
				// Best-effort set 500 if nothing meaningful has been written yet
				if recorder.statusCode < 400 {
					recorder.WriteHeader(http.StatusInternalServerError)
				}
			}

			// Collect metric information
			method := r.Method
			endpoint := m.normalize(r)
			statusCode := strconv.Itoa(recorder.statusCode)
			duration := time.Since(start).Seconds()

			// Update metrics
			m.processedOpsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
			m.requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)

			// Increment failed operations metric on error
			if recorder.statusCode >= 400 {
				errorType := "client_error"
				if recorder.statusCode >= 500 {
					errorType = "server_error"
				}
				m.failedOpsTotal.WithLabelValues(method, endpoint, errorType).Inc()
			}
		}()

		// Process the request
		next.ServeHTTP(recorder, r)
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

func (m *metrics) SetServiceUp() {
	m.serviceUp.Set(1)
}

// SetServiceDown marks service as down when shutting down
func (m *metrics) SetServiceDown() {
	m.serviceUp.Set(0)
}

// UpdateSystemMetrics updates system resource metrics
func (m *metrics) UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage
	m.memoryUsageBytes.Set(float64(memStats.Alloc))

	// Update goroutine count
	m.goroutinesCount.Set(float64(runtime.NumGoroutine()))
}

// RecordGCDuration records garbage collection duration
func (m *metrics) RecordGCDuration(duration time.Duration) {
	m.gcDuration.Observe(duration.Seconds())
}

func (m *metrics) IncreaseProcessedOps(method, endpoint, statusCode string) {
	m.processedOpsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

func (m *metrics) IncreaseFailedOps(method, endpoint, errorType string) {
	m.failedOpsTotal.WithLabelValues(method, endpoint, errorType).Inc()
}

func (m *metrics) ObserveRequestDuration(method, endpoint, statusCode string, duration float64) {
	m.requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
}

// defaultEndpointNormalize is the default normalizer that returns the raw path
func defaultEndpointNormalize(r *http.Request) string {
	return r.URL.Path
}
