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
	port := flag.Int("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	flag.Parse()

	log.Println("Starting RPS client...")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	out := com.Message{Value: "Ping"}
	if err := com.Write(conn, out); err != nil {
		log.Printf("Failed to write data: %s", err)
	}

	input, err := com.Read[com.Message](conn)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
	}

	log.Printf("Read message: %+v", input)
	log.Println("Successfully pinged server!")
}
