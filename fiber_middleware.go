//go:build fiber

package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PrometheusMiddleware returns a Fiber middleware function for Prometheus metrics
// This requires the fiber dependency and should be built with -tags fiber
func (m *Metrics) PrometheusMiddleware() fiber.Handler {
	// Mark service as up when it starts
	m.ServiceUp.Set(1)

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process the request
		err := c.Next()

		// Collect metric information
		method := c.Method()
		endpoint := c.Route().Path
		statusCode := strconv.Itoa(c.Response().StatusCode())
		duration := time.Since(start).Seconds()

		// Update metrics
		m.ProcessedOpsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		m.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)

		// Increment failed operations metric on error
		if c.Response().StatusCode() >= 400 {
			errorType := "client_error"
			if c.Response().StatusCode() >= 500 {
				errorType = "server_error"
			}
			m.FailedOpsTotal.WithLabelValues(method, endpoint, errorType).Inc()
		}

		return err
	}
}
