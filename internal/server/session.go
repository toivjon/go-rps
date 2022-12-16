package server

import (
	"fmt"
	"log"

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
	session := &Session{
		cli1:  cli1,
		cli2:  cli2,
		round: NewRound(),
	}
	cli1.session = session
	cli2.session = session
	return session
}

// Start starts the target session by notifying target clients to start the actual gaming.
func (s *Session) Start() error {
	if err := s.cli1.WriteStart(s.cli2.name); err != nil {
		return fmt.Errorf("failed to write START message for %s. %w", s.cli1, err)
	}
	if err := s.cli2.WriteStart(s.cli1.name); err != nil {
		return fmt.Errorf("failed to write START message for %s. %w", s.cli2, err)
	}
	log.Printf("Session %#p started (%s & %s)", s, s.cli1, s.cli2)
	return nil
}

// Select applies the given selection for the target client for the ongoing RPS game round.
func (s *Session) Select(cli *Client, selection game.Selection) error {
	switch cli {
	case s.cli1:
		s.round.selection1 = selection
	case s.cli2:
		s.round.selection2 = selection
	}
	if s.round.Ended() {
		result1, result2 := s.round.Result()
		if err := s.cli1.WriteResult(s.round.selection2, result1); err != nil {
			return fmt.Errorf("failed to write RESULT message for %s. %w", s.cli1, err)
		}
		if err := s.cli2.WriteResult(s.round.selection1, result2); err != nil {
			return fmt.Errorf("failed to write RESULT message for %s. %w", s.cli2, err)
		}
		log.Printf("Session %#p round result %s:%s and %s:%s", s, s.cli1, result1, s.cli2, result2)
		if result1 == game.ResultDraw && result2 == game.ResultDraw {
			s.round = NewRound()
		}
	}
	return nil
}

// Close closes the target session by removing session references and closing attached connections.
func (s *Session) Close() {
	s.cli1.session = nil
	s.cli2.session = nil
	log.Printf("Session %#p closed (conn1: %s conn2: %s)", s, s.cli1, s.cli2)
	s.cli1.Close()
	s.cli2.Close()
}
