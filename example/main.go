package main

import (
	"fmt"
	"log"
	"net/http"

	metrics "github.com/BiTaksi/prometheus-metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create metrics instance
	config := &metrics.MetricsConfig{
		Namespace: "example",
		Subsystem: "api",
	}
	m := metrics.NewMetrics(config)

	// Add business metrics
	m.AddBusinessCounter("user_registrations_total", "Total number of user registrations", []string{"source", "plan"})
	m.AddBusinessGauge("active_users", "Number of currently active users", []string{"region"})
	m.AddBusinessHistogram("order_value", "Order value distribution", []string{"currency"}, []float64{10, 50, 100, 500, 1000})

	// Create HTTP router
	mux := http.NewServeMux()

	// Main endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! This is a Prometheus metrics example.\n")
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Business metrics test endpoints
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		plan := r.URL.Query().Get("plan")

		if source == "" {
			source = "web"
		}
		if plan == "" {
			plan = "free"
		}

		// Update business metric
		m.IncrementBusinessCounter("user_registrations_total", source, plan)

		fmt.Fprintf(w, "User registration recorded for source: %s, plan: %s\n", source, plan)
	})

	mux.HandleFunc("/active-users", func(w http.ResponseWriter, r *http.Request) {
		region := r.URL.Query().Get("region")
		if region == "" {
			region = "us-east"
		}

		// Simulate active user count
		count := 150.0
		m.SetBusinessGauge("active_users", count, region)

		fmt.Fprintf(w, "Active users updated for region: %s, count: %.0f\n", region, count)
	})

	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		currency := r.URL.Query().Get("currency")
		if currency == "" {
			currency = "USD"
		}

		// Simulate order value
		value := 75.5
		m.ObserveBusinessHistogram("order_value", value, currency)

		fmt.Fprintf(w, "Order recorded for currency: %s, value: %.2f\n", currency, value)
	})

	// System metrics endpoint
	mux.HandleFunc("/system", func(w http.ResponseWriter, r *http.Request) {
		// Update system metrics
		m.UpdateSystemMetrics()
		fmt.Fprintf(w, "System metrics updated\n")
	})

	// Add metrics middleware
	handler := m.HTTPMiddleware(mux)

	// Metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", handler)

	fmt.Println("Server started: http://localhost:8080")
	fmt.Println("Metrics: http://localhost:8080/metrics")
	fmt.Println("Health: http://localhost:8080/health")
	fmt.Println("System metrics: http://localhost:8080/system")
	fmt.Println("Business Metrics:")
	fmt.Println("  Register: http://localhost:8080/register?source=web&plan=premium")
	fmt.Println("  Active Users: http://localhost:8080/active-users?region=eu-west")
	fmt.Println("  Order: http://localhost:8080/order?currency=EUR")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
