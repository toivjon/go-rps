package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

var ErrEnd = errors.New("end")

// State represents a reference to a client state which may return a next state or an error.
type State func(conn net.Conn) (State, error)

func Run(port uint, host string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	for state := Connected; state != nil; {
		if state, err = state(conn); err != nil && !errors.Is(err, ErrEnd) {
			return err
		}
	}
	return nil
}

func Connected(conn net.Conn) (State, error) {
	log.Printf("Enter your name:")
	name, err := waitInput()
	if err != nil {
		return nil, fmt.Errorf("failed to read user input to as username. %w", err)
	}
	if err := com.WriteMessage(conn, com.TypeJoin, com.JoinContent{Name: name}); err != nil {
		return nil, fmt.Errorf("failed to write JOIN message. %w", err)
	}
	log.Printf("Joined the game as %q.", name)
	return Joined, nil
}

func Joined(conn net.Conn) (State, error) {
	log.Printf("Waiting for an opponent. Please wait...")
	message, err := waitMessage[com.StartContent](conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read START message. %w", err)
	}
	log.Printf("Opponent %q joined the game.", message.OpponentName)
	return Started, nil
}

func Started(conn net.Conn) (State, error) {
	log.Println("Please type the selection ('r', 'p', 's') and press enter")
	selection, err := waitSelection()
	if err != nil {
		return nil, fmt.Errorf("failed to read selection. %w", err)
	}
	if err := com.WriteMessage(conn, com.TypeSelect, com.SelectContent{Selection: selection}); err != nil {
		return nil, fmt.Errorf("failed to write SELECT message. %w", err)
	}
	return Waiting, nil
}

func Waiting(conn net.Conn) (State, error) {
	log.Println("Waiting for game result. Please wait...")
	message, err := waitMessage[com.ResultContent](conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read RESULT message. %w", err)
	}
	log.Printf("Result: %+v", message)
	if message.Result == game.ResultDraw {
		log.Println("Round ended in a draw. Let's have an another round...")
		return Started, nil
	}
	return nil, ErrEnd
}

func waitMessage[T any](conn net.Conn) (*T, error) {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read message. %w", err)
	}
	content := new(T)
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message content. %w", err)
	}
	return content, nil
}

func waitInput() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to scan user input. %w", scanner.Err())
	}
	return scanner.Text(), nil
}

func waitSelection() (game.Selection, error) {
	input, err := waitInput()
	if err != nil {
		return "", fmt.Errorf("failed to scan user input for selection. %w", err)
	}
	selection := game.Selection(input)
	if err := game.ValidateSelection(selection); err != nil {
		return "", fmt.Errorf("failed to validate selection. %w", err)
	}
	return selection, nil
}
