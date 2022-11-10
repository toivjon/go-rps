package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

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

	start(conn, *name)
}

func start(conn net.Conn, name string) {
	inbox := newInbox(conn)
	for {
		select {
		case message := <-inbox:
			switch message.Type {
			case com.MessageTypeConnect:
				break
			case com.MessageTypeConnected:
				break
			}
			log.Printf("Received message: %+v", message)
			return
		case <-time.After(time.Second):
			if err := com.WriteConnect(conn, com.ConnectContent{Name: name}); err != nil {
				log.Printf("Failed to write data: %s", err)
			}
		}
	}
}

func newInbox(conn net.Conn) <-chan com.Message {
	inbox := make(chan com.Message)
	go func() {
		for {
			message, err := com.Read[com.Message](conn)
			if errors.Is(err, io.EOF) {
				log.Printf("Closing client. Server closed the connection.")
				break
			} else if err != nil {
				log.Printf("Failed to read message: %s", err)
				break
			}
			inbox <- *message
		}
	}()
	return inbox
}
