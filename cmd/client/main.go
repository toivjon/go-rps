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
	port := flag.Int("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	flag.Parse()

	log.Println("Starting RPS client...")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, bufferSize)
	if _, err = conn.Write([]byte("Ping")); err != nil {
		log.Printf("Failed to write buffer: %s", err)
		return
	}

	readByteCount, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read buffer: %s", err)
		return
	}

	log.Printf("Received %d bytes: %s", readByteCount, string(buffer[:readByteCount]))
	log.Println("Successfully pinged server!")
}
