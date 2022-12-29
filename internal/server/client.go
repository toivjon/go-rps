package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

// Client represents a single client connected to the server.
type Client struct {
	Conn    io.ReadWriteCloser
	Name    string
	Session *Session
}

// NewClient builds a new client with the provided connection.
func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{
		Conn:    conn,
		Name:    "",
		Session: nil,
	}
}

// WriteStart sends a START message to the client.
func (c *Client) WriteStart(opponentName string) error {
	content := com.StartContent{OpponentName: opponentName}
	if err := com.WriteMessage(c.Conn, com.TypeStart, content); err != nil {
		return fmt.Errorf("failed to write START message. %w", err)
	}
	return nil
}

// WriteResult sends a RESULT message to the client.
func (c *Client) WriteResult(opponentSelection game.Selection, result game.Result) error {
	messageContent := com.ResultContent{OpponentSelection: opponentSelection, Result: result}
	if err := com.WriteMessage(c.Conn, com.TypeResult, messageContent); err != nil {
		return fmt.Errorf("failed to write RESULT message. %w", err)
	}
	return nil
}

// Run starts the processing of the client.
func (c *Client) Run(
	leaveCh chan<- io.ReadWriteCloser,
	joinCh chan<- Message[com.JoinContent],
	selectCh chan<- Message[com.SelectContent],
) {
	defer func() {
		leaveCh <- c.Conn
		c.Conn.Close()
	}()
	for {
		message, err := com.Read[com.Message](c.Conn)
		if err != nil {
			return
		}
		switch message.Type {
		case com.TypeJoin:
			content := new(com.JoinContent)
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			joinCh <- Message[com.JoinContent]{Conn: c.Conn, Content: *content}
		case com.TypeSelect:
			content := new(com.SelectContent)
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			selectCh <- Message[com.SelectContent]{Conn: c.Conn, Content: *content}
		case com.TypeResult, com.TypeStart:
			log.Printf("Connection %#p received unsupported message type %s!", c.Conn, message.Type)
			return
		}
	}
}

// String returns a string representing the client.
func (c *Client) String() string {
	return fmt.Sprintf("client(%#p:%s)", c.Conn, c.Name)
}

// Close will close the client connection.
func (c *Client) Close() error {
	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("failed to close conn %#p. %w", c.Conn, err)
	}
	return nil
}
