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

var errConnClosed = errors.New("remote machine closed the connection")

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
	end := make(chan error)
	inbox := newInbox(conn, end)
	for {
		select {
		case err := <-end:
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
