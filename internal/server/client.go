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
	conn    io.ReadWriteCloser
	name    string
	session *Session
}

// NewClient builds a new client with the provided connection.
func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{
		conn:    conn,
		name:    "",
		session: nil,
	}
}

// WriteStart sends a START message to the client.
func (c *Client) WriteStart(opponentName string) error {
	content := com.StartContent{OpponentName: opponentName}
	if err := com.WriteMessage(c.conn, com.TypeStart, content); err != nil {
		return fmt.Errorf("failed to write START message. %w", err)
	}
	return nil
}

// WriteResult sends a RESULT message to the client.
func (c *Client) WriteResult(opponentSelection game.Selection, result game.Result) error {
	messageContent := com.ResultContent{OpponentSelection: opponentSelection, Result: result}
	if err := com.WriteMessage(c.conn, com.TypeResult, messageContent); err != nil {
		return fmt.Errorf("failed to write RESULT message. %w", err)
	}
	return nil
}

// String returns a string representing the client.
func (c *Client) String() string {
	return fmt.Sprintf("client(%#p:%s)", c.conn, c.name)
}

// Close will close the client connection.
func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close conn %#p. %w", c.conn, err)
	}
	return nil
}

// Run starts the processing of the client.
func (c *Client) Run(server *Server) {
	defer func() {
		server.leaveCh <- c.conn
		c.conn.Close()
	}()
	for {
		message, err := com.Read[com.Message](c.conn)
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
			server.joinCh <- Message[com.JoinContent]{conn: c.conn, content: *content}
		case com.TypeSelect:
			content := new(com.SelectContent)
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			server.selectCh <- Message[com.SelectContent]{conn: c.conn, content: *content}
		case com.TypeResult, com.TypeStart:
			log.Printf("Connection %#p received unsupported message type %s!", c.conn, message.Type)
			return
		}
	}
}
