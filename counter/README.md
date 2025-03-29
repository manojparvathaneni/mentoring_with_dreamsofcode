
# Building a Resilient Counter API in Go

> 🛠️ **Note:** This markdown file is a personal learning log — not production-ready code. It's intended to document my exploration of Go concepts while building a simple, evolving counter API. Each part builds on the previous one and captures lessons learned. The plan is to continue evolving this API and this documentation as I experiment with new ideas.

This guide walks you through the **evolution** of a counter API in Go. We begin with a basic goal and incrementally build toward a robust, concurrent, persistent API. Each stage includes a runnable contrived example focused on one concept, followed by an enhancement that moves us closer to the final project.

## 🎯 End Goal
Create a Go-based HTTP API that:
- Returns a visit count in JSON
- Is concurrency-safe
- Persists data to disk
- Works reliably across multiple instances (file locking)

---

## 📘 Table of Contents

- [🎯 End Goal](#-end-goal)
- [🔹 Step 1: In-Memory Counter with Basic Output](#-step-1-in-memory-counter-with-basic-output)
- [🔹 Step 2: Proper JSON Response](#-step-2-proper-json-response)
- [🔹 Step 3: Add Concurrency Safety with Mutex](#-step-3-add-concurrency-safety-with-mutex)
- [🔹 Step 4: Save Counter to a File](#-step-4-save-counter-to-a-file)
- [🔹 Step 5: Load Counter from File on Startup](#-step-5-load-counter-from-file-on-startup)
- [🔹 Step 6: Use `atomic.Int64` in a Struct and Return JSON](#-step-6-use-atomicint64-in-a-struct-and-return-json)
- [🔹 Step 8: File Locking with `syscall.Flock`](#-step-8-file-locking-with-syscallflock)
- [🏁 Final Implementation: Full API](#-final-implementation-full-api)
- [🧠 Summary of Concepts Learned](#-summary-of-concepts-learned


## 🪜 Step-by-Step Evolution

### 🔹 Step 1: In-Memory Counter with Basic Output
```go
package main

import (
	"fmt"
	"net/http"
)

var count int

func handler(w http.ResponseWriter, r *http.Request) {
	count++
	fmt.Fprintf(w, "Visits: %d", count)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
```
**Concepts Introduced:**
- HTTP handler
- Global state
- Simple response

---

### 🔹 Step 2: Proper JSON Response
```go
package main

import (
	"encoding/json"
	"net/http"
)

var count int

func handler(w http.ResponseWriter, r *http.Request) {
	count++
	json.NewEncoder(w).Encode(map[string]int{"visits": count})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
```
**New Concept:**
- Use of `encoding/json` for proper JSON API response

---

### 🔹 Step 3: Add Concurrency Safety with Mutex
```go
package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

var (
	count int
	mu    sync.Mutex
)

func handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	current := count
	mu.Unlock()
	json.NewEncoder(w).Encode(map[string]int{"visits": current})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
```
**Concepts Introduced:**
- `sync.Mutex` for safe concurrent access

---

### 🔹 Step 4: Save Counter to a File
```go
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

var (
	count int
	mu    sync.Mutex
)

func saveToFile(count int) {
	data, _ := json.Marshal(map[string]int{"visits": count})
	os.WriteFile("counter.json", data, 0644)
}

func handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	saveToFile(count)
	mu.Unlock()
	json.NewEncoder(w).Encode(map[string]int{"visits": count})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
```
**Concepts Introduced:**
- File persistence with `os.WriteFile`

---

### 🔹 Step 5: Load Counter from File on Startup
```go
func loadFromFile() int {
	b, err := os.ReadFile("counter.json")
	if err != nil {
		return 0
	}
	var d map[string]int
	json.Unmarshal(b, &d)
	return d["visits"]
}

func main() {
	count = loadFromFile()
	...
```
**Concepts Introduced:**
- JSON file reading & unmarshaling

---

### 🔹 Step 6: Use `atomic.Int64` Instead of Mutex
```go
import "sync/atomic"

type VisitCounter struct {
	Visits atomic.Int64
}

counter := &VisitCounter{}
counter.Visits.Add(1)
fmt.Println(counter.Visits.Load())
```
**Concepts Introduced:**
- `sync/atomic` for lock-free concurrency

---

### 🔹 Step 7: Serialize Atomic Counter with Helper Struct
```go
type visitData struct {
	Visits int64 `json:"visits"`
}

data := visitData{Visits: counter.Visits.Load()}
json.NewEncoder(w).Encode(data)
```
**Why?** `atomic.Int64` is not JSON-marshalable.

---

### 🔹 Step 8: File Locking with `syscall.Flock`
```go
import "syscall"

f, _ := os.OpenFile("counter.json", os.O_RDWR, 0644)
syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
// write
syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
```
**Concepts Introduced:**
- Prevent race conditions across multiple running instances using file-level locks

---

## 🏁 Final Implementation: Full API
```go
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

```
**What It Includes:**
- HTTP server
- Atomic counter
- JSON response
- Persistent file storage
- File locking (shared for read, exclusive for write)

---

## 🧠 Summary of Concepts Learned
- HTTP Handlers in Go
- JSON encoding with `encoding/json`
- Concurrency control with `sync.Mutex` and `sync/atomic`
- File I/O with `os.ReadFile`, `os.WriteFile`
- Struct marshaling and unmarshaling
- File locking with `syscall.Flock`

Each stage is an isolated, testable example that can be run independently and built upon. The end result is a well-structured, concurrent, persistent counter service you can run with multiple instances for resilience.
