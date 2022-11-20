package server

import (
	"net"

	"github.com/toivjon/go-rps/internal/game"
)

// Player represents a player client with a connection and a name.
type Player struct {
	Conn      net.Conn
	Name      string
	Selection chan game.Selection
	Finished  chan struct{}
}
