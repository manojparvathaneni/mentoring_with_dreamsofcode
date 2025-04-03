package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Ticker started. Press Ctrl+C to stop.")
	for i := 0; i < 5; i++ {
		<-ticker.C
		fmt.Println("Tick", i+1)
	}
}
