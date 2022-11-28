package server

import (
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

func runSession(matchmaker map[net.Conn]*Player, player *Player) {
	for opponentConn, opponent := range matchmaker {
		delete(matchmaker, opponent.Conn)
		log.Printf("Starting session %q vs %q", player.Name, opponent.Name)
		if err := com.WriteMessage(player.Conn, com.TypeStart, com.StartContent{OpponentName: opponent.Name}); err != nil {
			log.Panicln(err)
		}
		if err := com.WriteMessage(opponentConn, com.TypeStart, com.StartContent{OpponentName: player.Name}); err != nil {
			log.Panicln(err)
		}
		go runRound(player, opponent)
		log.Printf("Sent start message to %v and %v", player.Conn, opponent.Conn)
		return
	}
}
