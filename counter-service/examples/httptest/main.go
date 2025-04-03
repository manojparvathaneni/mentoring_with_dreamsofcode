package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Test response")
}

func main() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	fmt.Println("Status Code:", resp.StatusCode)
}
