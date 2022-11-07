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
	defaultName = "anonymous"
)

func main() {
	port := flag.Int("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	name := flag.String("name", defaultName, "The name of the player.")
	flag.Parse()

	log.Println("Starting RPS client...")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := com.WriteConnect(conn, com.ConnectContent{Name: *name}); err != nil {
		log.Printf("Failed to write data: %s", err)
	}

	input, err := com.Read[com.Message](conn)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
	}

	log.Printf("Read message: %+v", input)
	log.Println("Successfully pinged server!")
}
