package client

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

type WaitResult struct{}

func (w *WaitResult) Run(cli *Client) error {
	log.Println("Waiting for game result. Please wait...")
	select {
	case message := <-cli.inbox:
		return w.handleMessage(cli, message)
	case err := <-cli.disconnect:
		return w.handleDisconnect(err)
	}
}

func (w *WaitResult) handleMessage(cli *Client, message com.Message) error {
	if message.Type == com.TypeResult {
		content := new(com.ResultContent)
		if err := json.Unmarshal(message.Content, &content); err != nil {
			return fmt.Errorf("failed to read RESULT content. %w", err)
		}
		if content.Result == game.ResultDraw {
			log.Printf("Round ended in a draw. Let's have an another round...")
			return new(Round).Run(cli)
		}
		log.Printf("Result: %+v", content)
		return nil
	}
	return nil
}

func (w *WaitResult) handleDisconnect(err error) error {
	return nil
}
