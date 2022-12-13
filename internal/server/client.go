package server

import "net"

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
