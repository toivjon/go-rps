package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

	log.Println("Welcome to the RPS server")
	if err := run(*port, *host); err != nil {
		log.Fatalf("Server was closed due an error: %v", err)
	}
	log.Println("Server was closed successfully.")
}

func run(port uint, host string) error {
	log.Printf("Starting up server: %s:%d", host, port)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	server := server.NewServer(listener, shutdown)
	server.Run()
	return nil
}
