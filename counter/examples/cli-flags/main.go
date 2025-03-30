package main

import (
	"flag"
	"fmt"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	fmt.Println("Port:", *port)
	fmt.Println("Debug mode:", *debug)
}
