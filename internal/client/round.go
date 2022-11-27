package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

func playRound(conn net.Conn, inbox <-chan com.Message, disconnect chan error) error {
	log.Println("Please type the selection ('r', 'p', 's') and press enter:")
	select {
	case input := <-waitInput(disconnect):
		return handleInput(conn, inbox, disconnect, input)
	case message := <-inbox:
		return handlePlayRoundMessage(message)
	case err := <-disconnect:
		return err
	}
}

func handleInput(conn net.Conn, inbox <-chan com.Message, disconnect chan error, input string) error {
	selection := game.Selection(input)
	if err := game.ValidateSelection(selection); err != nil {
		return fmt.Errorf("failed to validate user input. %w", err)
	}
	if err := com.WriteMessage(conn, com.TypeSelect, com.SelectContent{Selection: selection}); err != nil {
		return fmt.Errorf("failed to write SELECT message. %w", err)
	}
	return waitResult(conn, inbox, disconnect)
}

func waitInput(disconnect chan error) <-chan string {
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

func handlePlayRoundMessage(message com.Message) error {
	// ... can occur if other player suddenly leaves the game?
	return nil
}
