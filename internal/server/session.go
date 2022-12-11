package server

import (
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
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

// Start starts the target session by notifying target clients to start the actual gaming.
func (s *Session) Start() error {
	if err := com.WriteMessage(s.cli1.conn, com.TypeStart, com.StartContent{OpponentName: s.cli2.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", s.cli1.conn, err)
	}
	if err := com.WriteMessage(s.cli2.conn, com.TypeStart, com.StartContent{OpponentName: s.cli1.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", s.cli2.conn, err)
	}
	s.cli1.session = s
	s.cli2.session = s
	log.Printf("Session %#p started (conn1: %#p conn2: %#p)", s, s.cli1.conn, s.cli2.conn)
	return nil
}

// Close closes the target session by removing session references and closing attached connections.
func (s *Session) Close() {
	s.cli1.session = nil
	s.cli2.session = nil
	log.Printf("Session %#p closed (conn1: %#p conn2: %#p)", s, &s.cli1.conn, &s.cli2.conn)
	s.cli1.conn.Close()
	s.cli2.conn.Close()
}
