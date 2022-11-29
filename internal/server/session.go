package server

import (
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
)

type session struct {
	player1 *Player
	player2 *Player
}

func (s *session) Start() error {
	log.Printf("Session %p starting...", s)
	if err := com.WriteMessage(s.player1.Conn, com.TypeStart, com.StartContent{OpponentName: s.player2.Name}); err != nil {
		return fmt.Errorf("failed to write start message for client %p. %w", s.player1, err)
	}
	if err := com.WriteMessage(s.player2.Conn, com.TypeStart, com.StartContent{OpponentName: s.player1.Name}); err != nil {
		return fmt.Errorf("failed to write start message for client %p. %w", s.player2, err)
	}
	go runRound(s.player1, s.player2)
	log.Printf("Session %p started.", s)
	return nil
}

func (s *session) Close() {
	log.Printf("Closing session %p: %q vs. %q", s, s.player1.Name, s.player2.Name)
	s.player1.Conn.Close()
	s.player2.Conn.Close()
	log.Printf("Closing session %p: %q vs. %q completed.", s, s.player1.Name, s.player2.Name)
}

func newSession(player1, player2 *Player) *session {
	session := &session{
		player1: player1,
		player2: player2,
	}
	player1.Session = session
	player2.Session = session
	return session
}
