package server

import (
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

// Session represents a single game session where to clients battle against each other in RPS rounds.
type Session struct {
	cli1  *Client
	cli2  *Client
	round *Round
}

// NewSession builds a new session for the given clients and attachs the session relation.
func NewSession(cli1, cli2 *Client) *Session {
	session := new(Session)
	session.cli1 = cli1
	session.cli2 = cli2
	session.round = NewRound(session)
	cli1.session = session
	cli2.session = session
	return session
}

// Start starts the target session by notifying target clients to start the actual gaming.
func (s *Session) Start() error {
	if err := com.WriteMessage(s.cli1.conn, com.TypeStart, com.StartContent{OpponentName: s.cli2.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", s.cli1.conn, err)
	}
	if err := com.WriteMessage(s.cli2.conn, com.TypeStart, com.StartContent{OpponentName: s.cli1.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", s.cli2.conn, err)
	}
	log.Printf("Session %#p started (conn1: %#p conn2: %#p)", s, s.cli1.conn, s.cli2.conn)
	return nil
}

// Select applies the given selection for the target client for the ongoing RPS game round.
func (s *Session) Select(cli *Client, selection game.Selection) {
	switch cli.conn {
	case s.cli1.conn:
		s.round.selection1 = selection
	case s.cli2.conn:
		s.round.selection2 = selection
	}
	if s.round.Ended() {
		result1, result2 := s.round.Result()
		messageContent := com.ResultContent{OpponentSelection: s.round.selection2, Result: result1}
		if err := com.WriteMessage(s.cli1.conn, com.TypeResult, messageContent); err != nil {
			log.Printf("Failed to write RESULT message for conn %#p. %s", s.cli1.conn, err)
			s.Close()
			return
		}
		messageContent = com.ResultContent{OpponentSelection: s.round.selection1, Result: result2}
		if err := com.WriteMessage(s.cli2.conn, com.TypeResult, messageContent); err != nil {
			log.Printf("failed to write RESULT message for conn  %#p. %s", s.cli2.conn, err)
			s.Close()
			return
		}
		log.Printf("Session %#p round result %#p:%s and %#p:%s", s, s.cli1.conn, result1, s.cli2.conn, result2)
		if result1 == game.ResultDraw && result2 == game.ResultDraw {
			s.round = NewRound(s)
		}
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
