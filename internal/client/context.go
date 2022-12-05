package client

import (
	"io"
)

// Context represents a client processing context.
type Context struct {
	Input io.Reader
	Conn  io.ReadWriter
}

// NewContext builds a new client context with the given input and connection.
func NewContext(input io.Reader, conn io.ReadWriter) Context {
	return Context{
		Input: input,
		Conn:  conn,
	}
}
