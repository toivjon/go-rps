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
	close := make(chan error)
	inbox := newInbox(conn, close)
	for {
		select {
		case err := <-close:
			log.Printf("Received close signal: %v", err)
			return
		case message := <-inbox:
			switch message.Type {
			case com.MessageTypeConnect:
				break
			case com.MessageTypeConnected:
				break
			}
			log.Printf("Received message: %+v", message)
		case <-time.After(time.Second):
			if err := com.WriteConnect(conn, com.ConnectContent{Name: name}); err != nil {
				log.Printf("Failed to write data: %s", err)
			}
		}
	}
}

func newInbox(conn net.Conn, close chan<- error) <-chan com.Message {
	inbox := make(chan com.Message)
	go func() {
		for {
			message, err := com.Read[com.Message](conn)
			switch {
			case errors.Is(err, io.EOF):
				close <- errors.New("remote machine closed the connection")
			case err != nil:
				close <- fmt.Errorf("failed to read message. %w", err)
			default:
				inbox <- *message
			}
		}
	}()
	return inbox
}
