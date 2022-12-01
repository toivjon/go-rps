package server

import (
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
)

type Session struct {
	player1 *Player
	player2 *Player
}

// NewSession builds a new session for the given players.
func NewSession(player1, player2 *Player) *Session {
	session := &Session{
		player1: player1,
		player2: player2,
	}
	player1.Session = session
	player2.Session = session
	return session
}

// Start begins the session by notifying the players and starting a game round.
func (s *Session) Start() error {
	log.Printf("Session %#p starting...", s)
	if err := com.WriteMessage(s.player1.Conn, com.TypeStart, com.StartContent{OpponentName: s.player2.Name}); err != nil {
		return fmt.Errorf("failed to write start message for client %#p. %w", s.player1, err)
	}
	if err := com.WriteMessage(s.player2.Conn, com.TypeStart, com.StartContent{OpponentName: s.player1.Name}); err != nil {
		return fmt.Errorf("failed to write start message for client %#p. %w", s.player2, err)
	}
	go runRound(s.player1, s.player2)
	log.Printf("Session %#p started.", s)
	return nil
}

// Close ends the session by notifying the players by closing the connections.
func (s *Session) Close() {
	log.Printf("Closing session %#p: %q vs. %q", s, s.player1.Name, s.player2.Name)
	s.player1.Conn.Close()
	s.player2.Conn.Close()
	log.Printf("Closing session %#p: %q vs. %q completed.", s, s.player1.Name, s.player2.Name)
}
