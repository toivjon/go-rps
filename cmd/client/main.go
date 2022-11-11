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

var (
	errConnClosed = errors.New("remote machine closed the connection")
	errConnected  = errors.New("received unsupported CONNECTED message")
)

func main() {
	port := flag.Uint("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	name := flag.String("name", defaultName, "The name of the player.")
	flag.Parse()

	log.Println("Starting RPS client...")
	if err := start(*port, *host, *name); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}

func start(port uint, host, name string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	end := make(chan error)      // A channel for ending the application.
	inbox := newInbox(conn, end) // A generator for incoming messages.

	for {
		select {
		case err := <-end:
			return err
		case message := <-inbox:
			switch message.Type {
			case com.MessageTypeConnected:
				// ... wait for start message.
				break
			case com.MessageTypeConnect:
				return errConnected
			}
			log.Printf("Received message: %+v", message)
		case <-time.After(time.Second):
			if err := com.WriteConnect(conn, com.ConnectContent{Name: name}); err != nil {
				log.Printf("Failed to write data: %s", err)
			}
		}
	}
}

func newInbox(conn net.Conn, end chan<- error) <-chan com.Message {
	inbox := make(chan com.Message)
	go func() {
		for {
			message, err := com.Read[com.Message](conn)
			switch {
			case errors.Is(err, io.EOF):
				end <- errConnClosed
			case err != nil:
				end <- fmt.Errorf("failed to read message. %w", err)
			default:
				inbox <- *message
			}
		}
	}()
	return inbox
}
