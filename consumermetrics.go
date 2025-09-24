package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type IConsumerMetrics interface {
	IBaseMetrics
	StartServer(port int) error
	IncMessagesConsumed()
	IncMessageErrors(errorType string)
	ObserveMessageProcessingDuration(duration float64)
}

type consumerMetrics struct {
	IBaseMetrics
	consumed prometheus.Counter
	errors   *prometheus.CounterVec
	duration prometheus.Histogram
}

func NewConsumerMetrics() IConsumerMetrics {
	var messagesConsumed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "consumer_messages_total",
			Help: "Sum of messages consumed",
		},
	)

	var messageErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consumer_message_errors_total",
			Help: "Sum of messages that could not be processed",
		},
		[]string{"error_type"},
	)

	var messageProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "consumer_message_processing_seconds",
			Help:    "Message processing durations",
			Buckets: prometheus.DefBuckets,
		},
	)

	m := &consumerMetrics{
		IBaseMetrics: initializeBaseMetrics(),
		consumed:     messagesConsumed,
		errors:       messageErrors,
		duration:     messageProcessingDuration,
	}

	prometheus.MustRegister(m.consumed, m.errors, m.duration)

	return m
}

func (p *consumerMetrics) StartServer(port int) error {
	http.Handle("/metrics", promhttp.Handler())

	errCh := make(chan error, 1)

	go func() {
		fmt.Println("Prometheus metrics server running on :2112")
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			errCh <- err
		}
	}()

	return <-errCh
}

func (p *consumerMetrics) IncMessagesConsumed() {
	p.consumed.Inc()
}

func (p *consumerMetrics) IncMessageErrors(errorType string) {
	p.errors.WithLabelValues(errorType).Inc()
}

func (p *consumerMetrics) ObserveMessageProcessingDuration(duration float64) {
	p.duration.Observe(duration)
}
