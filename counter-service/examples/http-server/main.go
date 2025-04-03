package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Current count: 42")
	})

	fmt.Println("Serving on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
