package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/counter"
	"github.com/yourusername/counter-service/internal/metrics"
)

// CreateTempFile creates a temporary file for testing
func CreateTempFile(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	dir, err := os.MkdirTemp("", "counter-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a path for the temporary file
	path := filepath.Join(dir, "counter-test.json")

	// Return the path and a cleanup function
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return path, cleanup
}

// NewTestConfig creates a config for testing
func NewTestConfig(t *testing.T) *config.Config {
	t.Helper()

	// Create a temporary file
	path, cleanup := CreateTempFile(t)
	t.Cleanup(cleanup)

	// Create a test config
	cfg := &config.Config{
		Port:              "8099",
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       5 * time.Second,
		ShutdownTimeout:   1 * time.Second,
		Filename:          path,
		FilePermissions:   0644,
		SaveRetryAttempts: 1,
		SaveRetryDelay:    10 * time.Millisecond,
		PersistInterval:   100 * time.Millisecond,
		RateLimit:         100,
		RateBurst:         200,
		EnableMetrics:     true,
		EnableCORS:        true,
		AllowedOrigins:    []string{"*"},
		LogLevel:          "fatal", // Silence logs during tests
		Environment:       "test",
	}

	return cfg
}

// NewTestLogger creates a logger for testing
func NewTestLogger() *zerolog.Logger {
	// Create a silent logger for tests
	logger := zerolog.New(io.Discard).With().Timestamp().Logger()
	return &logger
}

// NewTestMetrics creates metrics for testing
func NewTestMetrics() *metrics.Metrics {
	return metrics.NewMetrics()
}

// NewTestCounterService creates a counter service for testing
func NewTestCounterService(t *testing.T) *counter.Service {
	t.Helper()

	cfg := NewTestConfig(t)
	logger := NewTestLogger()
	metrics := NewTestMetrics()

	// Create a test counter service
	service, err := counter.NewService(cfg, logger, metrics)
	if err != nil {
		t.Fatalf("Failed to create counter service: %v", err)
	}

	// Clean up the service when the test is done
	t.Cleanup(func() {
		service.Shutdown()
	})

	return service
}

// PerformRequest performs an HTTP request against a handler for testing
func PerformRequest(t *testing.T, method, path string, body interface{}, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Record response
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	return w
}