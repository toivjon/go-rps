package main

import (
	"flag"
	"log"

	"github.com/toivjon/go-rps/internal/client"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

func main() {
	port := flag.Uint("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	flag.Parse()

	log.Println("Welcome to the RPS client")
	if err := client.Run(*port, *host); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}
