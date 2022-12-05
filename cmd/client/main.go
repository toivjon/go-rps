package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

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
	if err := run(*port, *host); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}

func run(port uint, host string) error {
	log.Printf("Connecting to server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection. %w", err)
	}
	defer conn.Close()
	if err := client.Run(client.NewContext(os.Stdin, conn), client.Connected); err != nil {
		return fmt.Errorf("failed to run client. %w", err)
	}
	return nil
}
