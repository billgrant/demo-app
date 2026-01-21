package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics
// These are package-level so handlers and middleware can update them

var (
	// httpRequestsTotal counts all HTTP requests
	// Labels let us slice by method (GET/POST), path (/api/items), and status (200/404/500)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "demoapp_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration tracks response time distribution
	// Histogram automatically creates buckets (0.005s, 0.01s, 0.025s, ... 10s)
	// Labels: method and path (not status, since we don't know status until response)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "demoapp_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // Default: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		},
		[]string{"method", "path"},
	)

	// itemsTotal is a gauge showing current item count
	// Gauge because it can go up (create) or down (delete)
	itemsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "demoapp_items_total",
			Help: "Current number of items in the database",
		},
	)

	// displayUpdatesTotal counts POST requests to /api/display
	// Counter because it only increases
	displayUpdatesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "demoapp_display_updates_total",
			Help: "Total number of display panel updates",
		},
	)

	// buildInfo is a gauge that's always 1, with labels for version info
	// This is a common Prometheus pattern for exposing build metadata
	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "demoapp_info",
			Help: "Build information (always 1)",
		},
		[]string{"version"},
	)
)

// init registers all metrics with Prometheus
// init() runs automatically before main() — Go calls it for every file that has one
func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(itemsTotal)
	prometheus.MustRegister(displayUpdatesTotal)
	prometheus.MustRegister(buildInfo)

	// Set build info (always 1, labels carry the metadata)
	// TODO: Set version from build flags in CI/CD
	buildInfo.WithLabelValues("dev").Set(1)
}
