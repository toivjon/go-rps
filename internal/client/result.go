package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

var errInvalidMessageType = errors.New("invalid message type")

func waitResult(conn net.Conn, inbox <-chan com.Message, disconnect chan error) error {
	log.Printf("Waiting for game result. Please wait...")
	select {
	case message := <-inbox:
		return handleWaitResultMessage(conn, inbox, disconnect, message)
	case err := <-disconnect:
		return err
	}
}

func handleWaitResultMessage(
	conn net.Conn,
	inbox <-chan com.Message,
	disconnect chan error,
	message com.Message,
) error {
	switch message.Type {
	case com.TypeResult:
		content := new(com.ResultContent)
		if err := json.Unmarshal(message.Content, &content); err != nil {
			return fmt.Errorf("failed to read RESULT content. %w", err)
		}
		if content.Result == game.ResultDraw {
			log.Printf("Round ended in a draw. Let's have an another round...")
			return playRound(conn, inbox, disconnect)
		}
		log.Printf("Result: %+v", content)
		return nil
	case com.TypeJoin, com.TypeSelect, com.TypeStart:
		fallthrough
	default:
		return errInvalidMessageType
	}
}
