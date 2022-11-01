package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

func main() {
	port := flag.Int("port", defaultPort, "The port to listen for incoming TCP connections.")
	flag.Parse()

	log.Println("Starting RPS server...")

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", defaultHost, *port))
	if err != nil {
		log.Fatalf("Unable to start listening TCP socket on port %d: %s", *port, err.Error())
	}
	defer server.Close()
	log.Printf("Waiting for clients on port: %d", *port)

	for {
		_, err := server.Accept()
		if err != nil {
			log.Printf("Error accepting: %s", err.Error())
		} else {
			log.Printf("Client connected!")
			// ... handle connection with a new Go routine.
			break
		}
	}
}
