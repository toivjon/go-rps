package server

import (
	"log"

	"github.com/toivjon/go-rps/internal/com"
)

type session struct {
	player1 *Player
	player2 *Player
}

func runSession(session *session) {
	player1 := session.player1
	player2 := session.player2
	log.Printf("Starting session %q vs %q", player1.Name, player2.Name)
	if err := com.WriteMessage(player1.Conn, com.TypeStart, com.StartContent{OpponentName: player2.Name}); err != nil {
		log.Panicln(err)
	}
	if err := com.WriteMessage(player2.Conn, com.TypeStart, com.StartContent{OpponentName: player1.Name}); err != nil {
		log.Panicln(err)
	}
	go runRound(player1, player2)
	log.Printf("Sent start message to %v and %v", player1.Conn, player2.Conn)
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

func endSession(session *session) {
	session.player1.Conn.Close()
	session.player2.Conn.Close()
}
