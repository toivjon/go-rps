package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Welcome to the RPS server")
	server := server.NewServer()
	if err := server.Run(*port, *host, shutdown); err != nil {
		log.Fatalf("Server was closed due an error: %v", err)
	}
	log.Println("Server was closed successfully.")
}
