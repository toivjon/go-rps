package server

import (
	"encoding/json"
	"fmt"
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

func processConnection(disconnect chan<- net.Conn, player *Player, matchmaker *matchmaker) {
	defer func() {
		disconnect <- player.Conn
		player.Conn.Close()
	}()

	joinContent, err := readJoin(player.Conn)
	if err != nil {
		return
	}
	player.Name = joinContent.Name

	if err := matchmaker.Enter(player); err != nil {
		log.Printf("Player %q Failed to enter matchmaker. %s", player.Name, err)
	} else {
		log.Printf("Player %q joined.", player.Name)
	}

	reader := func() chan game.Selection {
		outbox := make(chan game.Selection)
		go func() {
			for {
				selection, err := readSelect(player.Conn)
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

func readJoin(conn net.Conn) (com.JoinContent, error) {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		return com.JoinContent{}, fmt.Errorf("failed to read JOIN message. %w", err)
	}
	content := com.JoinContent{Name: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.JoinContent{}, fmt.Errorf("failed to read JOIN content. %w", err)
	}
	return content, nil
}

func readSelect(conn net.Conn) (com.SelectContent, error) {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		return com.SelectContent{}, fmt.Errorf("failed to read SELECT message. %w", err)
	}
	content := com.SelectContent{Selection: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.SelectContent{}, fmt.Errorf("failed to read SELECT content. %w", err)
	}
	return content, nil
}
