package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

type Round struct{}

func (r *Round) Run(cli *Client) error {
	log.Println("Please type the selection ('r', 'p', 's') and press enter:")
	select {
	case input := <-r.waitInput(cli.disconnect):
		return r.handleInput(cli, input)
	case message := <-cli.inbox:
		return r.handleMessage(message)
	case err := <-cli.disconnect:
		return r.handleDisconnect(err)
	}
}

func (r *Round) handleMessage(message com.Message) error {
	return nil
}

func (r *Round) handleDisconnect(err error) error {
	// ... log something here? e.g. Server closed the connection or such?
	return nil
}

func (r *Round) handleInput(cli *Client, input string) error {
	selection := game.Selection(input)
	if err := game.ValidateSelection(selection); err != nil {
		return fmt.Errorf("failed to validate user input. %w", err)
	}
	content, err := json.Marshal(com.SelectContent{Selection: selection})
	if err != nil {
		return fmt.Errorf("failed marshal SELECT content into JSON. %w", err)
	}
	if err := com.Write(cli.conn, com.Message{Type: com.TypeSelect, Content: content}); err != nil {
		return fmt.Errorf("failed to write SELECT message to connection. %w", err)
	}
	return new(WaitResult).Run(cli)
}

func (r *Round) waitInput(disconnect chan<- error) <-chan string {
	input := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			disconnect <- fmt.Errorf("failed to scan user input. %w", scanner.Err())
		}
		input <- scanner.Text()
	}()
	return input
}
