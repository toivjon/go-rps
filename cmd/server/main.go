package main

import (
	"log"
	"net"
)

const (
	ServerHost = "localhost"
	ServerPort = "7777"
)

func main() {
	log.Println("Starting RPS server...")

	server, err := net.Listen("tcp", ServerHost+":"+ServerPort)
	if err != nil {
		log.Fatalf("Unable to start listening TCP socket on port %s: %s", ServerPort, err.Error())
	}
	defer server.Close()
	log.Printf("Waiting for clients on port: %s", ServerPort)

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
