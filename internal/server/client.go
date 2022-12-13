package server

import (
	"encoding/json"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

// Client represents a single client connected to the server.
type Client struct {
	conn    net.Conn
	name    string
	session *Session
}

// NewClient builds a new client with the provided connection.
func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:    conn,
		name:    "",
		session: nil,
	}
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
