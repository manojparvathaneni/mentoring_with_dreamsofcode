package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("Program is running. Press Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutdown signal received. Cleaning up...")
	time.Sleep(2 * time.Second)
	fmt.Println("Shutdown complete.")
}
