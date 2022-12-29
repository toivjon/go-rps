package server

import (
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/game"
)

// Session represents a single game session where to clients battle against each other in RPS rounds.
type Session struct {
	Cli1  *Client
	Cli2  *Client
	Round *Round
}

// NewSession builds a new session for the given clients and attachs the session relation.
func NewSession(cli1, cli2 *Client) *Session {
	session := &Session{
		Cli1:  cli1,
		Cli2:  cli2,
		Round: NewRound(),
	}
	cli1.Session = session
	cli2.Session = session
	return session
}

// Start starts the target session by notifying target clients to start the actual gaming.
func (s *Session) Start() error {
	if err := s.Cli1.WriteStart(s.Cli2.Name); err != nil {
		return fmt.Errorf("failed to write START message for %s. %w", s.Cli1, err)
	}
	if err := s.Cli2.WriteStart(s.Cli1.Name); err != nil {
		return fmt.Errorf("failed to write START message for %s. %w", s.Cli2, err)
	}
	log.Printf("Session %#p started (%s & %s)", s, s.Cli1, s.Cli2)
	return nil
}

// Select applies the given selection for the target client for the ongoing RPS game round.
func (s *Session) Select(cli *Client, selection game.Selection) error {
	switch cli {
	case s.Cli1:
		s.Round.Selection1 = selection
	case s.Cli2:
		s.Round.Selection2 = selection
	}
	if s.Round.Ended() {
		result1, result2 := s.Round.Result()
		if err := s.Cli1.WriteResult(s.Round.Selection2, result1); err != nil {
			return fmt.Errorf("failed to write RESULT message for %s. %w", s.Cli1, err)
		}
		if err := s.Cli2.WriteResult(s.Round.Selection1, result2); err != nil {
			return fmt.Errorf("failed to write RESULT message for %s. %w", s.Cli2, err)
		}
		log.Printf("Session %#p round result %s:%s and %s:%s", s, s.Cli1, result1, s.Cli2, result2)
		if result1 == game.ResultDraw && result2 == game.ResultDraw {
			s.Round = NewRound()
		}
	}
	return nil
}

// Close closes the target session by removing session references and closing attached connections.
func (s *Session) Close() {
	s.Cli1.Session = nil
	s.Cli2.Session = nil
	log.Printf("Session %#p closed (%s & %s)", s, s.Cli1, s.Cli2)
	s.Cli1.Close()
	s.Cli2.Close()
}
