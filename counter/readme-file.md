# Counter Service

A production-ready visit counter service with robust persistence, metrics, and high performance.

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Learning Path](#learning-path)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Persistent Counter**: Maintains visit count with disk persistence
- **High Performance**: Optimized for high throughput with atomic operations
- **Configuration**: Environment-based configuration management
- **Metrics**: Prometheus integration for observability
- **Logging**: Structured logging with request tracing
- **API**: RESTful API with standardized responses
- **Security**: Rate limiting and CORS support
- **Reliability**: Atomic file operations with CRC verification
- **Graceful Shutdown**: Ensures data persistence on shutdown

## Installation

### Prerequisites

- Go 1.18 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/counter-service.git
cd counter-service

# Build the binary
go build -o counter-service

# Run the service
./counter-service
```

### Using Docker

```bash
# Build the Docker image
docker build -t counter-service .

# Run the container
docker run -p 8090:8090 counter-service
```

## Configuration

The service can be configured via:

1. Configuration file (`config.yaml`)
2. Environment variables
3. Command-line flags

### Configuration File Example

Create a `config.yaml` file in the project directory:

```yaml
port: "8090"
filename: "counter.json"
shutdownTimeout: 10s
persistInterval: 5m
logLevel: "info"
enableMetrics: true
enableCORS: true
allowedOrigins:
  - "http://localhost:3000"
  - "https://yourdomain.com"
environment: "production"
```

### Environment Variables

All settings can be overridden using environment variables with the prefix `COUNTER_`:

```bash
COUNTER_PORT=9000 COUNTER_LOGLEVEL=debug ./counter-service
```

### Available Settings

| Setting | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| port | COUNTER_PORT | 8090 | Server port |
| filename | COUNTER_FILENAME | counter.json | Data storage file |
| shutdownTimeout | COUNTER_SHUTDOWNTIMEOUT | 10s | Graceful shutdown timeout |
| persistInterval | COUNTER_PERSISTINTERVAL | 5m | Background persistence interval |
| logLevel | COUNTER_LOGLEVEL | info | Log level (debug, info, warn, error) |
| enableMetrics | COUNTER_ENABLEMETRICS | true | Enable Prometheus metrics |
| enableCORS | COUNTER_ENABLECORS | true | Enable CORS support |
| allowedOrigins | COUNTER_ALLOWEDORIGINS | * | Comma-separated list of allowed origins |
| environment | COUNTER_ENVIRONMENT | development | Environment (development, production) |

## API Reference

### Increment Counter

```
POST /api/counter/increment
```

Increments the counter and returns the new value.

**Response Example:**

```json
{
  "success": true,
  "data": {
    "visits": 42
  },
  "request_id": "1647359121-1",
  "response_time_ms": 0.523
}
```

### Get Counter

```
GET /api/counter
```

Returns the current counter value without incrementing.

**Response Example:**

```json
{
  "success": true,
  "data": {
    "visits": 42
  },
  "request_id": "1647359122-2",
  "response_time_ms": 0.128
}
```

### Health Check

```
GET /health
```

Returns the service health status.

**Response Example:**

```json
{
  "success": true,
  "data": {
    "status": "UP",
    "timestamp": "2025-03-31T14:22:56Z",
    "version": "1.0.0",
    "buildInfo": {
      "goVersion": "go1.18.3",
      "platform": "linux/amd64"
    }
  },
  "request_id": "1647359123-3",
  "response_time_ms": 0.051
}
```

### Metrics

```
GET /metrics
```

Returns Prometheus metrics for monitoring.

## Learning Path

Follow this step-by-step guide to master the concepts implemented in this project:

### 1. Understanding Go Concurrency (Week 1)

**Concepts to Learn:**
- Goroutines
- Channels
- Atomic operations
- Mutexes and locks

**Practice Exercise:**
Create a simple in-memory counter with concurrent access:

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// Using atomic operations
	var atomicCounter int64
	
	// Using mutex
	var mutexCounter int
	var mutex sync.Mutex
	
	var wg sync.WaitGroup
	
	// Launch 1000 goroutines
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Increment atomic counter
			atomic.AddInt64(&atomicCounter, 1)
			
			// Increment mutex-protected counter
			mutex.Lock()
			mutexCounter++
			mutex.Unlock()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Atomic counter: %d\n", atomicCounter)
	fmt.Printf("Mutex counter: %d\n", mutexCounter)
}
```

### 2. File Operations in Go (Week 2)

**Concepts to Learn:**
- File I/O in Go
- Error handling
- JSON serialization
- File locking

**Practice Exercise:**
Create a basic persistent counter that saves to a file:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

type Counter struct {
	Value int `json:"value"`
}

func main() {
	// Load counter
	counter, err := loadCounter("counter.json")
	if err != nil {
		fmt.Println("Error loading counter:", err)
		counter = &Counter{Value: 0}
	}
	
	// Increment counter
	counter.Value++
	fmt.Println("Counter value:", counter.Value)
	
	// Save counter
	if err := saveCounter("counter.json", counter); err != nil {
		fmt.Println("Error saving counter:", err)
	}
}

func loadCounter(filename string) (*Counter, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	// Apply shared lock
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	
	if len(data) == 0 {
		return &Counter{Value: 0}, nil
	}
	
	var counter Counter
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, err
	}
	
	return &counter, nil
}

func saveCounter(filename string, counter *Counter) error {
	data, err := json.Marshal(counter)
	if err != nil {
		return err
	}
	
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	// Apply exclusive lock
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	
	if _, err := f.Write(data); err != nil {
		return err
	}
	
	return f.Sync()
}
```

### 3. HTTP Servers in Go (Week 3)

**Concepts to Learn:**
- HTTP handlers
- Routing
- Middleware
- Response writing

**Practice Exercise:**
Build a basic HTTP server with handlers:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Hello, World!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	
	// Apply middleware
	var handler http.Handler = mux
	handler = loggingMiddleware(handler)
	
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	
	log.Println("Server starting on port 8080")
	log.Fatal(server.ListenAndServe())
}
```

### 4. Configuration Management (Week 4)

**Concepts to Learn:**
- Environment variables
- Configuration files
- Using Viper

**Practice Exercise:**
Implement configuration with Viper:

```go
package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port       string
	DataFile   string
	LogLevel   string
	Production bool
}

func loadConfig() Config {
	// Set defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("dataFile", "data.json")
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("production", false)
	
	// Look for config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	
	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error reading config file: %v", err)
		}
	}
	
	// Create config struct
	config := Config{
		Port:       viper.GetString("port"),
		DataFile:   viper.GetString("dataFile"),
		LogLevel:   viper.GetString("logLevel"),
		Production: viper.GetBool("production"),
	}
	
	fmt.Printf("Loaded configuration: %+v\n", config)
	return config
}

func main() {
	config := loadConfig()
	// Use config values to start your application
}
```

### 5. Structured Logging (Week 4)

**Concepts to Learn:**
- Structured vs traditional logging
- Log levels
- Using zerolog

**Practice Exercise:**
Implement structured logging with zerolog:

```go
package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogging(production bool) {
	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	
	// Configure logger
	if !production {
		// Pretty console logging for development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON logging for production
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func main() {
	setupLogging(false)
	
	// Different log levels
	log.Debug().Str("component", "main").Msg("This is a debug message")
	log.Info().Str("component", "main").Msg("Server starting up")
	log.Warn().Str("component", "main").Int("retries", 3).Msg("Connection unstable")
	log.Error().Str("component", "main").Err(os.ErrNotExist).Msg("File not found")
	
	// Structured logging with fields
	log.Info().
		Str("user", "john").
		Int("id", 123).
		Bool("admin", true).
		Msg("User logged in")
}
```

### 6. Metrics and Monitoring (Week 5)

**Concepts to Learn:**
- Prometheus metrics types
- Metric collection
- Exposing metrics

**Practice Exercise:**
Implement Prometheus metrics:

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter metric
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	// Gauge metric
	connectionCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_connections_current",
			Help: "Current number of connections",
		},
	)
	
	// Histogram metric
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Start timer
	timer := prometheus.NewTimer(requestDuration.WithLabelValues(r.URL.Path))
	defer timer.ObserveDuration()
	
	// Increment connection counter
	connectionCount.Inc()
	defer connectionCount.Dec()
	
	// Simulate processing
	time.Sleep(100 * time.Millisecond)
	
	// Increment request counter
	requestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	
	w.Write([]byte("Hello, world!"))
}

func main() {
	http.HandleFunc("/hello", handleRequest)
	http.Handle("/metrics", promhttp.Handler())
	
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 7. Rate Limiting (Week 5)

**Concepts to Learn:**
- Rate limiting algorithms
- Token bucket
- Using golang.org/x/time/rate

**Practice Exercise:**
Implement rate limiting middleware:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/time/rate"
)

type rateLimitMiddleware struct {
	limiter *rate.Limiter
	next    http.Handler
}

func newRateLimitMiddleware(rps int, burst int, next http.Handler) *rateLimitMiddleware {
	return &rateLimitMiddleware{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
		next:    next,
	}
}

func (m *rateLimitMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.limiter.Allow() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		response := map[string]string{
			"error": "Rate limit exceeded",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	m.next.ServeHTTP(w, r)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func main() {
	// Create handler
	handler := http.HandlerFunc(helloHandler)
	
	// Wrap with rate limiting middleware (10 requests per second with burst of 20)
	rateLimited := newRateLimitMiddleware(10, 20, handler)
	
	// Register handler
	http.Handle("/hello", rateLimited)
	
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 8. Graceful Shutdown (Week 6)

**Concepts to Learn:**
- Signal handling
- Context cancellation
- HTTP server shutdown

**Practice Exercise:**
Implement a graceful shutdown:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create server
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(helloHandler),
	}
	
	// Start server in a goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutdown signal received")
	
	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Server has been gracefully shut down")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second) // Simulate work
	w.Write([]byte("Hello, World!"))
}
```

### 9. CORS Configuration (Week 6)

**Concepts to Learn:**
- Cross-Origin Resource Sharing
- CORS headers
- Using rs/cors package

**Practice Exercise:**
Implement CORS support:

```go
package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Hello, World!"}`))
}

func main() {
	// Create a new CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		Debug:          true,
	})
	
	// Create a handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", helloHandler)
	
	// Wrap with CORS middleware
	handler := corsMiddleware.Handler(mux)
	
	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### 10. Data Integrity (Week 7)

**Concepts to Learn:**
- CRC checksums
- Atomic file operations
- Error handling strategies

**Practice Exercise:**
Implement atomic file write with CRC:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Data struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	CRC     uint32 `json:"crc,omitempty"`
}

// calculateCRC computes a simple checksum
func calculateCRC(data []byte) uint32 {
	var crc uint32 = 0
	for _, b := range data {
		crc = crc*31 + uint32(b)
	}
	return crc
}

// atomicWriteFile writes data atomically using a temporary file
func atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory
	dir := filepath.Dir(filename)
	tempFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	
	// Clean up on any error
	defer func() {
		if err != nil {
			os.Remove(tempPath)
		}
		tempFile.Close()
	}()
	
	// Write data to temp file
	if _, err = tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	
	// Sync to ensure data is written to disk
	if err = tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	
	// Close the file before rename
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	
	// Rename the temp file (atomic on most filesystems)
	if err = os.Rename(tempPath, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

func saveData(filename string, data Data) error {
	// Create a copy for CRC calculation
	dataCopy := data
	dataCopy.CRC = 0
	
	// Marshal without CRC
	jsonBytes, err := json.MarshalIndent(dataCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	// Calculate CRC
	crc := calculateCRC(jsonBytes)
	
	// Add CRC to original data
	data.CRC = crc
	
	// Marshal with CRC
	jsonBytes, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data with CRC: %w", err)
	}
	
	// Write atomically
	if err := atomicWriteFile(filename, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

func loadData(filename string) (Data, error) {
	var data Data
	
	// Read file
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		return data, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Unmarshal data
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return data, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	
	// Verify CRC
	expectedCRC := data.CRC
	dataCopy := data
	dataCopy.CRC = 0
	
	// Marshal without CRC for verification
	jsonBytes, err = json.MarshalIndent(dataCopy, "", "  ")
	if err != nil {
		return data, fmt.Errorf("failed to marshal data for CRC check: %w", err)
	}
	
	calculatedCRC := calculateCRC(jsonBytes)
	if calculatedCRC != expectedCRC {
		return data, fmt.Errorf("CRC mismatch: expected %d, got %d", expectedCRC, calculatedCRC)
	}
	
	return data, nil
}

func main() {
	data := Data{
		ID:      1,
		Name:    "Test Data",
		Version: 1,
	}
	
	// Save data
	if err := saveData("data.json", data); err != nil {
		fmt.Printf("Error saving data: %v\n", err)
		return
	}
	
	// Load data
	loadedData, err := loadData("data.json")
	if err != nil {
		fmt.Printf("Error loading data: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded data: %+v\n", loadedData)
}
```

### 11. Building the Complete Application (Week 8)

**Concepts to Learn:**
- Project structure
- Dependency injection
- Integration testing

**Project Structure:**

```
counter-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   └── server.go
│   ├── config/
│   │   └── config.go
│   ├── counter/
│   │   ├── counter.go
│   │   ├── persistence.go
│   │   └── service.go
│   ├── metrics/
│   │   └── metrics.go
│   └── test/
│       └── helpers.go
├── pkg/
│   ├── fileutils/
│   │   └── fileutils.go
│   └── logging/
│       └── logging.go
├── .dockerignore
├── .gitignore
├── config.yaml
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

**Final Learning Exercise:**
Build a load testing tool for your counter service:

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Configuration
	baseURL := "http://localhost:8090"
	concurrentUsers := 10
	requestsPerUser := 100
	
	// Results channels
	results := make(chan time.Duration, concurrentUsers*requestsPerUser)
	errors := make(chan error, concurrentUsers*requestsPerUser)
	
	// Synchronization
	var wg sync.WaitGroup
	
	// Start time
	startTime := time.Now()
	
	// Start concurrent users
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Each user makes requestsPerUser requests
			for j := 0; j < requestsPerUser; j++ {
				// Send request
				start := time.Now()
				resp, err := http.Post(baseURL+"/api/counter/increment", "application/json", nil)
				duration := time.Since(start)
				
				if err != nil {
					errors <- fmt.Errorf("user %d, request %d: %w", userID, j, err)
					continue
				}
				
				// Check response
				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("user %d, request %d: got status %d", userID, j, resp.StatusCode)
					resp.Body.Close()
					continue
				}
				
				resp.Body.Close()
				results <- duration
			}
		}(i)
	}
	
	// Wait for all users to complete
	wg.Wait()
	close(results)
	close(errors)
	
	// Calculate statistics
	var totalDuration time.Duration
	var maxDuration time.Duration
	var count int
	
	for duration := range results {
		totalDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
		count++
	}
	
	// Print results
	totalTime := time.Since(startTime)
	avgDuration := totalDuration / time.Duration(count)
	requestsPerSecond := float64(count) / totalTime.Seconds()
	
	fmt.Printf("Load Test Results:\n")
	fmt.Printf("  Total requests:   %d\n", count)
	fmt.Printf("  Total errors:     %d\n", len(errors))
	fmt.Printf("  Total time:       %v\n", totalTime)
	fmt.Printf("  Average latency:  %v\n", avgDuration)
	fmt.Printf("  Max latency:      %v\n", maxDuration)
	fmt.Printf("  Requests/second:  %.2f\n", requestsPerSecond)
	
	// Print sample of errors
	errorCount := 0
	for err := range errors {
		if errorCount < 10 {
			fmt.Printf("Error: %v\n", err)
		}
		errorCount++
	}
	if errorCount > 10 {
		fmt.Printf("... and %d more errors\n", errorCount-10)
	}
}
```

## Additional Learning Resources

- [Go Documentation](https://golang.org/doc/)
- [Go By Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Concurrency Patterns](https://blog.golang.org/pipelines)
- [Google's Go Style Guide](https://google.github.io/styleguide/go/)
- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)
- [Zerolog Documentation](https://github.com/rs/zerolog)
- [Viper Documentation](https://github.com/spf13/viper)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
