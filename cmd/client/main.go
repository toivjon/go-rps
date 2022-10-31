package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

const (
	ServerHost = "localhost"
)

const defaultPort = 7777

func main() {
	port := flag.Int("port", defaultPort, "The port of the server TCP socket.")
	flag.Parse()

	log.Println("Starting RPS client...")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ServerHost, *port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	log.Println("Successfully pinged server!")
}
