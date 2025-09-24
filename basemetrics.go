package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type IBaseMetrics interface {
	AddBusinessCounter(name, help string, labels []string) *prometheus.CounterVec
	AddBusinessGauge(name, help string, labels []string) *prometheus.GaugeVec
	AddBusinessHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec
	GetBusinessCounter(name string) (*prometheus.CounterVec, bool)
	GetBusinessGauge(name string) (*prometheus.GaugeVec, bool)
	GetBusinessHistogram(name string) (*prometheus.HistogramVec, bool)
	IncrementBusinessCounter(name string, labelValues ...string)
	SetBusinessGauge(name string, value float64, labelValues ...string)
	ObserveBusinessHistogram(name string, value float64, labelValues ...string)
}

type baseMetrics struct {
	// Business metrics - extensible counters and gauges
	BusinessCounters   map[string]*prometheus.CounterVec
	BusinessGauges     map[string]*prometheus.GaugeVec
	BusinessHistograms map[string]*prometheus.HistogramVec
}

func initializeBaseMetrics() *baseMetrics {
	return &baseMetrics{
		BusinessCounters:   make(map[string]*prometheus.CounterVec),
		BusinessGauges:     make(map[string]*prometheus.GaugeVec),
		BusinessHistograms: make(map[string]*prometheus.HistogramVec),
	}
}

// AddBusinessCounter creates a new business counter metric
func (m *baseMetrics) AddBusinessCounter(name, help string, labels []string) *prometheus.CounterVec {
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
func (m *baseMetrics) AddBusinessGauge(name, help string, labels []string) *prometheus.GaugeVec {
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
func (m *baseMetrics) AddBusinessHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
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
func (m *baseMetrics) GetBusinessCounter(name string) (*prometheus.CounterVec, bool) {
	counter, exists := m.BusinessCounters[name]
	return counter, exists
}

// GetBusinessGauge returns a business gauge by name
func (m *baseMetrics) GetBusinessGauge(name string) (*prometheus.GaugeVec, bool) {
	gauge, exists := m.BusinessGauges[name]
	return gauge, exists
}

// GetBusinessHistogram returns a business histogram by name
func (m *baseMetrics) GetBusinessHistogram(name string) (*prometheus.HistogramVec, bool) {
	histogram, exists := m.BusinessHistograms[name]
	return histogram, exists
}

// IncrementBusinessCounter increments a business counter
func (m *baseMetrics) IncrementBusinessCounter(name string, labelValues ...string) {
	if counter, exists := m.BusinessCounters[name]; exists {
		counter.WithLabelValues(labelValues...).Inc()
	}
}

// SetBusinessGauge sets a business gauge value
func (m *baseMetrics) SetBusinessGauge(name string, value float64, labelValues ...string) {
	if gauge, exists := m.BusinessGauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Set(value)
	}
}

// ObserveBusinessHistogram observes a value in business histogram
func (m *baseMetrics) ObserveBusinessHistogram(name string, value float64, labelValues ...string) {
	if histogram, exists := m.BusinessHistograms[name]; exists {
		histogram.WithLabelValues(labelValues...).Observe(value)
	}
}
