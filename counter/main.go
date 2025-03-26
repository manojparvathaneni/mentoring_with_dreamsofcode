package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	Counter  map[string]int
	mu       sync.Mutex
	filename = "counter.json"
)

func init() {
	Counter = readCounter(filename)

	// Initialize if key not present
	if _, ok := Counter["visited"]; !ok {
		Counter["visited"] = 0
	}
}

func numOfVisits(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	Counter["visited"]++
	visitedCount := Counter["visited"]

	// Make a copy of the current counter to persist safely
	currentState := make(map[string]int)
	for k, v := range Counter {
		currentState[k] = v
	}
	mu.Unlock()

	// Persist the updated counter
	persistCounter(currentState)

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]int{"visited": visitedCount}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func persistCounter(counter map[string]int) {
	mu.Lock()
	defer mu.Unlock()

	jsonData, err := json.MarshalIndent(counter, "", "  ")
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		log.Println("Error writing JSON to file:", err)
	}
}

func readCounter(filename string) map[string]int {
	readData, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Counter file not found. Starting fresh.")
			return make(map[string]int)
		}
		log.Println("Error reading counter file:", err)
		return make(map[string]int)
	}

	var readMap map[string]int
	if err := json.Unmarshal(readData, &readMap); err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return make(map[string]int)
	}

	return readMap
}

func main() {
	http.HandleFunc("/api/counter", numOfVisits)

	log.Println("Starting Counter Server....")
	log.Println("Listening on port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
