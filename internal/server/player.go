package server

import "net"

// Player represents a player client with a connection and a name.
type Player struct {
	Conn net.Conn
	Name string
}
