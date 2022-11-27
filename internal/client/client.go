package client

import (
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

func Run(port uint, host, name string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	log.Printf("Joining as player: %s", name)
	if err := com.WriteMessage(conn, com.TypeJoin, com.JoinContent{Name: name}); err != nil {
		return fmt.Errorf("failed to write JOIN message. %w", err)
	}

	disconnect := make(chan error)
	inbox := inbox(conn, disconnect)
	if err := waitStart(conn, inbox, disconnect); err != nil {
		return fmt.Errorf("an error occurred during running the current state. %w", err)
	}
	log.Printf("Game over.")
	return nil
}

func inbox(conn net.Conn, disconnect chan<- error) <-chan com.Message {
	inbox := make(chan com.Message)
	go func() {
		for {
			message, err := com.Read[com.Message](conn)
			if err != nil {
				disconnect <- err
				break
			}
			inbox <- *message
		}
	}()
	return inbox
}
