package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"syscall"
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
	http.HandleFunc("/api/counter", updateCounter)

	log.Println("Starting Counter Server....")
	log.Println("Listening on port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
