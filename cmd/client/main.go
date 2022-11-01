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
const defaultHost = "localhost"

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

	log.Println("Successfully pinged server!")
}
