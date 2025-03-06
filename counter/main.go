package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

var (
	Counter = make(map[string]int)
	mu      sync.Mutex
)

func numOfVisits(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	Counter["visited"]++
	visitedCount := Counter["visited"]
	mu.Unlock()

	response := map[string]int{"visited": visitedCount}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func main() {
	Counter["visited"] = 0
	http.HandleFunc("/api/counter", numOfVisits)

	log.Println("Staring Counter Server....")
	log.Println("Listening on port 8090")
	http.ListenAndServe(":8090", nil)

}
