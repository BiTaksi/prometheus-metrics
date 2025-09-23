//go:build fasthttp

package metrics

import (
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

// FastHTTPMiddleware returns a FastHTTP middleware function for Prometheus metrics
// This requires the fasthttp dependency and should be built with -tags fasthttp
func (m *Metrics) FastHTTPMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	// Mark service as up when it starts
	m.ServiceUp.Set(1)

	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()

		// Process the request
		next(ctx)

		// Collect metric information
		method := string(ctx.Method())
		endpoint := string(ctx.Path())
		statusCode := strconv.Itoa(ctx.Response.StatusCode())
		duration := time.Since(start).Seconds()

		// Update metrics
		m.ProcessedOpsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		m.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)

		// Increment failed operations metric on error
		if ctx.Response.StatusCode() >= 400 {
			errorType := "client_error"
			if ctx.Response.StatusCode() >= 500 {
				errorType = "server_error"
			}
			m.FailedOpsTotal.WithLabelValues(method, endpoint, errorType).Inc()
		}
	}
}

// FastHTTPHandler creates a FastHTTP handler wrapper for easier integration
func (m *Metrics) FastHTTPHandler(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return m.FastHTTPMiddleware(handler)
}
