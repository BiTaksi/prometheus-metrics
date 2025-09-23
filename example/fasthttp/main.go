package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	metrics "github.com/BiTaksi/prometheus-metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func main() {
	// Create metrics instance
	config := &metrics.MetricsConfig{
		Namespace: "fasthttp_example",
		Subsystem: "api",
	}
	m := metrics.NewMetrics(config)

	// Main handler
	handler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			ctx.SetContentType("text/plain")
			fmt.Fprintf(ctx, "FastHTTP Prometheus metrics example!\n")
		case "/health":
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetContentType("text/plain")
			fmt.Fprintf(ctx, "OK")
		case "/system":
			// Update system metrics
			m.UpdateSystemMetrics()
			ctx.SetContentType("text/plain")
			fmt.Fprintf(ctx, "System metrics updated\n")
		case "/metrics":
			// Prometheus metrics endpoint (using HTTP adapter)
			fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(ctx)
		default:
			ctx.Error("Not found", fasthttp.StatusNotFound)
		}
	}

	// Add metrics middleware - inline implementation
	wrappedHandler := func(ctx *fasthttp.RequestCtx) {
		start := time.Now()

		// Mark service as up when it starts
		m.ServiceUp.Set(1)

		// Process the request
		handler(ctx)

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

	fmt.Println("FastHTTP Server started: http://localhost:8080")
	fmt.Println("Metrics: http://localhost:8080/metrics")
	fmt.Println("Health: http://localhost:8080/health")
	fmt.Println("System metrics: http://localhost:8080/system")

	log.Fatal(fasthttp.ListenAndServe(":8080", wrappedHandler))
}
