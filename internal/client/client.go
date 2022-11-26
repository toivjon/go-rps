package client

import (
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

type Client struct {
	conn       net.Conn
	inbox      <-chan com.Message
	disconnect chan error
}

func NewClient(conn net.Conn) *Client {
	disconnect := make(chan error)
	return &Client{
		conn:       conn,
		inbox:      inbox(conn, disconnect),
		disconnect: disconnect,
	}
}

func (c *Client) Run() error {
	err := new(WaitStart).Run(c)
	if err != nil {
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
