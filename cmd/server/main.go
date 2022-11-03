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

	bufferSize = 1024
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

	buffer := make([]byte, bufferSize)
	readByteCount, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read buffer: %s", err)
		return
	}

	log.Printf("Received %d bytes: %s", readByteCount, string(buffer[:readByteCount]))

	if _, err = conn.Write([]byte("Pong")); err != nil {
		log.Printf("Failed to write buffer: %s", err)
	}

	log.Printf("Successfully sent response.")
}
