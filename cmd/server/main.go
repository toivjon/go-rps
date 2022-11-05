package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

func main() {
	port := flag.Int("port", defaultPort, "The port to listen for connections.")
	host := flag.String("host", defaultHost, "The network address to listen for connections.")
	flag.Parse()

	log.Println("Starting RPS server...")

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Unable to start listening TCP socket on port %d: %s", *port, err.Error())
	}
	defer server.Close()
	log.Printf("Waiting for clients on port: %d", *port)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Error accepting: %s", err.Error())
		} else {
			log.Printf("Client connected!")
			go processConnection(conn)
		}
	}
}

func processConnection(conn net.Conn) {
	defer conn.Close()

	input, err := com.Read[com.Message](conn)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
		return
	}

	log.Printf("Read message: %+v", input)

	out := com.Message{Value: "Pong"}
	if err := com.Write(conn, out); err != nil {
		log.Printf("Failed to write data: %s", err)
		return
	}

	log.Printf("Successfully sent response.")
}
