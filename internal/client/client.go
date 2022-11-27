package client

import (
	"encoding/json"
	"fmt"
	"io"
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

	cli := NewClient(conn, name)
	if err := cli.Run(); err != nil {
		return fmt.Errorf("failed to run client. %w", err)
	}
	return nil
}

type Client struct {
	conn       net.Conn
	inbox      <-chan com.Message
	disconnect chan error
	name       string
}

func NewClient(conn net.Conn, name string) *Client {
	disconnect := make(chan error)
	return &Client{
		conn:       conn,
		inbox:      inbox(conn, disconnect),
		disconnect: disconnect,
		name:       name,
	}
}

func (c *Client) Run() error {
	log.Printf("Joining as player: %s", c.name)
	if err := send(c.conn, com.TypeJoin, com.JoinContent{Name: c.name}); err != nil {
		return err
	}

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

func send[T any](writer io.Writer, messageType com.MessageType, val T) error {
	content, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed marshal %s content into JSON. %w", messageType, err)
	}
	if err := com.Write(writer, com.Message{Type: messageType, Content: content}); err != nil {
		return fmt.Errorf("failed to write %s message to connection. %w", messageType, err)
	}
	return nil
}
