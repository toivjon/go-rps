package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/client"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
	defaultName = "anonymous"
)

func main() {
	port := flag.Uint("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	name := flag.String("name", defaultName, "The name of the player.")
	flag.Parse()

	log.Println("Welcome to the RPS client")
	if err := start(*port, *host, *name); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}

func start(port uint, host, name string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	cli := client.NewClient(conn, name)
	if err := cli.Run(); err != nil {
		return fmt.Errorf("failed to run client. %w", err)
	}
	return nil
}
