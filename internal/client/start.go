package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

func waitStart(conn net.Conn, inbox <-chan com.Message, disconnect chan error) error {
	log.Printf("Waiting for an opponent. Please wait...")
	select {
	case message := <-inbox:
		return handleWaitStartMessage(conn, inbox, disconnect, message)
	case err := <-disconnect:
		return err
	}
}

func handleWaitStartMessage(conn net.Conn, inbox <-chan com.Message, disconnect chan error, message com.Message) error {
	switch message.Type {
	case com.TypeStart:
		content := new(com.StartContent)
		if err := json.Unmarshal(message.Content, &content); err != nil {
			return fmt.Errorf("failed to read START content. %w", err)
		}
		log.Printf("Opponent joined the game: %s", content.OpponentName)
		return playRound(conn, inbox, disconnect)
	case com.TypeJoin, com.TypeResult, com.TypeSelect:
		fallthrough
	default:
		return errInvalidMessageType
	}
}
