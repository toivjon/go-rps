package main

import (
	"flag"
	"log"

	"github.com/toivjon/go-rps/internal/server"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

func main() {
	port := flag.Uint("port", defaultPort, "The port to listen for connections.")
	host := flag.String("host", defaultHost, "The network address to listen for connections.")
	flag.Parse()

	log.Println("Welcome to the RPS server")
	if err := server.Run(*port, *host); err != nil {
		log.Fatalf("Server was closed due an error: %v", err)
	}
	log.Println("Server was closed successfully.")
}
