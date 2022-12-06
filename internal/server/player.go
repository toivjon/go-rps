package server

import (
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

// Player represents a player client with a connection and a name.
type Player struct {
	Conn      net.Conn
	Name      string
	Selection chan game.Selection
	Finished  chan struct{}
	Session   *Session
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{
		Conn:      conn,
		Name:      "",
		Selection: make(chan game.Selection),
		Finished:  make(chan struct{}),
		Session:   nil,
	}
}

func processConnection(disconnect, join chan<- net.Conn, player *Player) {
	defer func() {
		disconnect <- player.Conn
		player.Conn.Close()
	}()

	joinContent, err := com.ReadMessage[com.JoinContent](player.Conn)
	if err != nil {
		return
	}
	player.Name = joinContent.Name

	join <- player.Conn

	reader := func() chan game.Selection {
		outbox := make(chan game.Selection)
		go func() {
			for {
				selection, err := com.ReadMessage[com.SelectContent](player.Conn)
				if err != nil {
					disconnect <- player.Conn
					return
				}
				outbox <- selection.Selection
			}
		}()
		return outbox
	}()

	for {
		log.Printf("Waiting for player %q selection or finished signal...", player.Name)
		select {
		case selection := <-reader:
			log.Printf("Received player %q selection: %s", player.Name, selection)
			player.Selection <- selection
		case <-player.Finished:
			return
		}
	}
}
