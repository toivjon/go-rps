package server

import (
	"log"

	"github.com/toivjon/go-rps/internal/game"
)

// Session represents a single game session where to clients battle against each other in RPS rounds.
type Session struct {
	cli1          *Client
	cli2          *Client
	cli1Selection game.Selection
	cli2Selection game.Selection
}

// NewSession builds a new session for  the given clients.
func NewSession(cli1, cli2 *Client) *Session {
	return &Session{
		cli1:          cli1,
		cli2:          cli2,
		cli1Selection: game.SelectionNone,
		cli2Selection: game.SelectionNone,
	}
}

// Close closes the target session by removing session references and closing attached connections.
func (s *Session) Close() {
	s.cli1.session = nil
	s.cli2.session = nil
	log.Printf("Session %#p closed (conn1: %#p conn2: %#p)", s, &s.cli1.conn, &s.cli2.conn)
	s.cli1.conn.Close()
	s.cli2.conn.Close()
}
