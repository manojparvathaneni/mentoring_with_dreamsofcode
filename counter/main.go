package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	Counter  = make(map[string]int)
	mu       sync.Mutex
	filename = "counter.json"
)

func init() {
	Counter = readCounter(filename)
}

func numOfVisits(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	Counter["visited"]++
	visitedCount := Counter["visited"]
	mu.Unlock()

	response := map[string]int{"visited": visitedCount}
	persistCounter(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func persistCounter(counter map[string]int) {
	jsonData, err := json.MarshalIndent(counter, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}
}

func readCounter(filename string) (counter map[string]int) {
	// Read JSON data from file
	readData, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return nil
	}

	// Convert JSON back to map
	var readMap map[string]int
	err = json.Unmarshal(readData, &readMap)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}
	return readMap
}

func main() {
	Counter["visited"] = 0
	http.HandleFunc("/api/counter", numOfVisits)

	log.Println("Staring Counter Server....")
	log.Println("Listening on port 8090")
	http.ListenAndServe(":8090", nil)

}
