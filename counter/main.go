package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type VisitCounter struct {
	Visits atomic.Int64
}

type counterData struct {
	Visits int64 `json:"visits"`
}

var (
	filename = "counter.json"
	counter  *VisitCounter
)

func init() {
	var err error
	counter, err = loadCounter()
	if err != nil {
		log.Println("Error loading counter:", err)
		counter = &VisitCounter{} // fallback to zero
	}
}

func saveCounter(counter *VisitCounter) error {
	data := counterData{Visits: counter.Visits.Load()}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Apply exclusive lock for writing
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	_, err = f.Write(jsonBytes)
	return err
}

// setupGracefulShutdown configures proper server shutdown
func setupGracefulShutdown(srv *http.Server) {
	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	go func() {
		<-stop

		log.Println("Server is shutting down...")

		// Save counter state before shutting down
		if err := saveCounter(counter); err != nil {
			log.Printf("Error saving counter during shutdown: %v", err)
		}

		// Create a deadline for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		log.Println("Server shutdown complete")
	}()
}

func loadCounter() (*VisitCounter, error) {
	counter := &VisitCounter{}

	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Apply shared lock for reading
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	var data counterData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		// If file is empty, return zero counter
		return counter, nil
	}

	counter.Visits.Store(data.Visits)
	return counter, nil
}

func updateCounter(w http.ResponseWriter, req *http.Request) {
	counter.Visits.Add(1)

	if err := saveCounter(counter); err != nil {
		log.Println("Error saving counter:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(counterData{Visits: counter.Visits.Load()}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {

	// Create a custom server mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/counter", updateCounter)

	// Configure the HTTP server
	server := &http.Server{
		Addr:           ":8090",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Set up graceful shutdown
	setupGracefulShutdown(server)

	// Start the server
	log.Printf("Starting Counter Server on port %s...", "8090")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed to start: %v", err)
	}
}
