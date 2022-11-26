package client

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
)

type WaitStart struct{}

func (s *WaitStart) Run(cli *Client) error {
	select {
	case message := <-cli.inbox:
		return s.handleMessage(cli, message)
	case err := <-cli.disconnect:
		return s.handleDisconnect(err)
	}
}

func (s *WaitStart) handleMessage(cli *Client, message com.Message) error {
	if message.Type == com.TypeStart {
		content := new(com.StartContent)
		if err := json.Unmarshal(message.Content, &content); err != nil {
			return fmt.Errorf("failed to read START content. %w", err)
		}
		log.Printf("Opponent joined the game: %s", content.OpponentName)
		return new(Round).Run(cli)
	}
	return nil // error?
}

func (s *WaitStart) handleDisconnect(err error) error {
	// ... log something here? e.g. Server closed the connection or such?
	return nil
}
