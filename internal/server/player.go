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
	Session   *session
}

func processConnection(conn net.Conn, disconnect chan<- net.Conn, player *Player, matchmaker *matchmaker) {
	defer func() {
		disconnect <- conn
		conn.Close()
	}()

	joinContent, err := readJoin(conn)
	if err != nil {
		disconnect <- conn
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
				selection, err := readSelect(conn)
				if err != nil {
					disconnect <- conn
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
