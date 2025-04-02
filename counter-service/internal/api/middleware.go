package api

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/metrics"
	"golang.org/x/time/rate"
)

// contextKey is a type for context keys
type contextKey string

// requestIDKey is the context key for request ID
const requestIDKey = contextKey("requestID")

// requestCounter is used to generate unique request IDs
var requestCounter int64

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// requestLogMiddleware logs HTTP requests
func requestLogMiddleware(logger *zerolog.Logger, metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), atomic.AddInt64(&requestCounter, 1))

			// Add request ID to context
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			rw := newResponseWriter(w)

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)
			durationSeconds := float64(duration) / float64(time.Second)

			// Update metrics
			metrics.RequestDuration.WithLabelValues(r.URL.Path).Observe(durationSeconds)
			metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", rw.status)).Inc()

			// Log request
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote", r.RemoteAddr).
				Int("status", rw.status).
				Str("requestID", requestID).
				Float64("duration_ms", float64(duration.Microseconds())/1000.0).
				Msg("Request processed")
		})
	}
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(logger *zerolog.Logger, limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if rate limit exceeded
			if !limiter.Allow() {
				logger.Warn().
					Str("remote", r.RemoteAddr).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("Rate limit exceeded")

				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// recoverMiddleware recovers from panics
func recoverMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error().
						Str("panic", fmt.Sprintf("%v", err)).
						Str("path", r.URL.Path).
						Msg("Recovered from panic")

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// metricsMiddleware captures basic metrics for each request
func metricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start the timer
			timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues(r.URL.Path))
			defer timer.ObserveDuration()

			next.ServeHTTP(w, r)
		})
	}
}