# Prometheus Metrics Package

A Go library that enables easy usage of Prometheus metrics in your microservices.

## Features

- HTTP request metrics (count, duration, error rate)
- System resource metrics (memory, goroutine count, GC duration)
- Flexible business metrics system (counter, gauge, histogram)
- Standard HTTP, Fiber and FastHTTP middleware support
- Namespace and subsystem support

## Installation

```bash
go get github.com/BiTaksi/prometheus-metrics
```

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/BiTaksi/prometheus-metrics"
)

func main() {
    // Create metrics instance
    config := &metrics.MetricsConfig{
        Namespace: "myservice",
        Subsystem: "api",
    }
    m := metrics.NewMetrics(config)
    
    // Use HTTP middleware
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Users endpoint"))
    })
    
    // Add metrics middleware
    handler := m.HTTPMiddleware(mux)
    
    // Add metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    http.Handle("/", handler)
    
    http.ListenAndServe(":8080", nil)
}
```

### Fiber Usage

To use Fiber middleware, build with `-tags fiber`:

```bash
go build -tags fiber
```

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/BiTaksi/prometheus-metrics"
)

func main() {
    app := fiber.New()
    
    // Create metrics instance
    config := &metrics.MetricsConfig{
        Namespace: "myservice",
        Subsystem: "api",
    }
    m := metrics.NewMetrics(config)
    
    // Add Prometheus middleware
    app.Use(m.PrometheusMiddleware())
    
    // Routes
    app.Get("/api/users", func(c *fiber.Ctx) error {
        return c.SendString("Users endpoint")
    })
    
    // Metrics endpoint (adapter required)
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
    
    app.Listen(":8080")
}
```

### Business Metrics

```go
// Create business metrics
m.AddBusinessCounter("user_registrations_total", "Total user registrations", []string{"source", "plan"})
m.AddBusinessGauge("active_users", "Currently active users", []string{"region"})
m.AddBusinessHistogram("order_value", "Order value distribution", []string{"currency"}, nil)

// Usage
m.IncrementBusinessCounter("user_registrations_total", "web", "premium")
m.SetBusinessGauge("active_users", 150, "us-east")
m.ObserveBusinessHistogram("order_value", 75.5, "USD")

// Get metric reference
if counter, exists := m.GetBusinessCounter("user_registrations_total"); exists {
    counter.WithLabelValues("mobile", "free").Inc()
}
```

### Manual Metric Usage

```go
// Update system metrics
m.UpdateSystemMetrics()

// Record GC duration
duration := time.Millisecond * 100
m.RecordGCDuration(duration)

// Mark service as down
m.SetServiceDown()
```

### FastHTTP Usage

To use FastHTTP middleware, build with `-tags fasthttp`:

```bash
go build -tags fasthttp
```

```go
package main

import (
    "github.com/valyala/fasthttp"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/valyala/fasthttp/fasthttpadaptor"
    "github.com/BiTaksi/prometheus-metrics"
)

func main() {
    // Create metrics instance
    config := &metrics.MetricsConfig{
        Namespace: "myservice",
        Subsystem: "api",
    }
    m := metrics.NewMetrics(config)
    
    // Main handler
    handler := func(ctx *fasthttp.RequestCtx) {
        switch string(ctx.Path()) {
        case "/api/users":
            ctx.SetContentType("text/plain")
            fmt.Fprintf(ctx, "Users endpoint")
        case "/metrics":
            // Prometheus metrics endpoint
            fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(ctx)
        default:
            ctx.Error("Not found", fasthttp.StatusNotFound)
        }
    }
    
    // Add metrics middleware
    wrappedHandler := m.FastHTTPHandler(handler)
    
    fasthttp.ListenAndServe(":8080", wrappedHandler)
}
```

### Configuration

```go
config := &metrics.MetricsConfig{
    Namespace: "myservice",  // Optional: metric prefix
    Subsystem: "api",       // Optional: metric subsystem
}

// Default configuration (empty namespace/subsystem)
config := metrics.DefaultConfig()
```

## Available Metrics

### HTTP Metrics
- `http_requests_total` - Total number of HTTP requests
- `http_requests_failed_total` - Total number of failed HTTP requests
- `http_request_duration_seconds` - HTTP request durations

### System Metrics
- `service_up` - Whether the service is up or down
- `memory_usage_bytes` - Memory usage
- `goroutines_count` - Number of active goroutines
- `gc_duration_seconds` - Garbage collection durations

## Example Projects

### Standard HTTP Example
Find a simple HTTP server example in the `example/` folder:

```bash
cd example
go run main.go
```

### FastHTTP Example
Find a FastHTTP server example in the `example/fasthttp/` folder:

```bash
cd example/fasthttp
go run -tags fasthttp main.go
```

Then visit http://localhost:8080/metrics to see the metrics.

## License

MIT