package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

type VisitCounter struct {
	Visits atomic.Int64
}

type counterData struct {
	Visits int64 `json:"visits"`
}

var (
	mu       sync.Mutex
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
	mu.Lock()
	defer mu.Unlock()

	data := counterData{Visits: counter.Visits.Load()}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonBytes, 0644)
}

func loadCounter() (*VisitCounter, error) {
	mu.Lock()
	defer mu.Unlock()

	counter := &VisitCounter{}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return counter, nil
		}
		return nil, err
	}

	var data counterData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
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
