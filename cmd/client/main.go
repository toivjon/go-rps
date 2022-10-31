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
	log.Println("Starting RPS client...")

	conn, err := net.Dial("tcp", ServerHost+":"+ServerPort)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	log.Println("Successfully pinged server!")
}
