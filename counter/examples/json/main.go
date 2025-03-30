package main

import (
	"encoding/json"
	"fmt"
)

type Counter struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func main() {
	original := Counter{Name: "page_views", Count: 42}
	data, err := json.Marshal(original)
	if err != nil {
		panic(err)
	}
	fmt.Println("JSON:", string(data))

	var decoded Counter
	if err := json.Unmarshal(data, &decoded); err != nil {
		panic(err)
	}
	fmt.Printf("Decoded struct: %+v\n", decoded)
}
