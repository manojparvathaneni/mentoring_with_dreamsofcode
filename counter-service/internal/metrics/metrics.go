package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus metrics for the application
type Metrics struct {
	// RequestsTotal counts the total number of HTTP requests
	RequestsTotal *prometheus.CounterVec

	// RequestDuration measures the duration of HTTP requests
	RequestDuration *prometheus.HistogramVec

	// CounterOperations counts counter operations by type
	CounterOperations *prometheus.CounterVec

	// CounterValue is the current value of the counter
	CounterValue prometheus.Gauge

	// OperationDuration measures the duration of counter operations
	OperationDuration *prometheus.HistogramVec

	// PersistErrors counts errors during persistence operations
	PersistErrors prometheus.Counter
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	// Create metrics
	metrics := &Metrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "counter_requests_total",
			Help: "The total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),

		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "counter_request_duration_seconds",
			Help:    "The duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"endpoint"}),

		CounterOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "counter_operations_total",
			Help: "The total number of counter operations",
		}, []string{"operation"}),

		CounterValue: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "counter_current_value",
			Help: "The current value of the counter",
		}),

		OperationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "counter_operation_duration_seconds",
			Help:    "Duration of counter operations in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"operation"}),

		PersistErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "counter_persist_errors_total",
			Help: "Total number of errors during counter persistence",
		}),
	}

	return metrics
}