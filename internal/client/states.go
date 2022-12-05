package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

var (
	ErrEnd         = errors.New("end")
	ErrNameTooLong = fmt.Errorf("player name must not contain more than %d characters", NameMaxLength)
)

// NameMaxLength specifies the maximum length of the player's name.
const NameMaxLength = 64

// State represents a reference to a client state which may return a next state or an error.
type State func(ctx Context) (State, error)

func Run(ctx Context, state State) error {
	for state != nil {
		nextState, err := state(ctx)
		if err != nil && !errors.Is(err, ErrEnd) {
			return err
		}
		state = nextState
	}
	return nil
}

func Connected(ctx Context) (State, error) {
	log.Printf("Enter your name:")
	name, err := waitInput(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read user input to as username. %w", err)
	}
	if len(name) > NameMaxLength {
		return nil, ErrNameTooLong
	}
	if err := com.WriteMessage(ctx.Conn, com.TypeJoin, com.JoinContent{Name: name}); err != nil {
		return nil, fmt.Errorf("failed to write JOIN message. %w", err)
	}
	log.Printf("Joined the game as %q.", name)
	return Joined, nil
}

func Joined(ctx Context) (State, error) {
	log.Printf("Waiting for an opponent. Please wait...")
	message, err := waitMessage[com.StartContent](ctx.Conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read START message. %w", err)
	}
	log.Printf("Opponent %q joined the game.", message.OpponentName)
	return Started, nil
}

func Started(ctx Context) (State, error) {
	log.Println("Please type the selection ('r', 'p', 's') and press enter")
	selection, err := waitSelection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read selection. %w", err)
	}
	if err := com.WriteMessage(ctx.Conn, com.TypeSelect, com.SelectContent{Selection: selection}); err != nil {
		return nil, fmt.Errorf("failed to write SELECT message. %w", err)
	}
	return Waiting, nil
}

func Waiting(ctx Context) (State, error) {
	log.Println("Waiting for game result. Please wait...")
	message, err := waitMessage[com.ResultContent](ctx.Conn)
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

func waitMessage[T any](conn io.ReadWriter) (*T, error) {
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

func waitInput(ctx Context) (string, error) {
	scanner := bufio.NewScanner(ctx.Input)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to scan user input. %w", scanner.Err())
	}
	return scanner.Text(), nil
}

func waitSelection(ctx Context) (game.Selection, error) {
	input, err := waitInput(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to scan user input for selection. %w", err)
	}
	selection := game.Selection(input)
	if err := game.ValidateSelection(selection); err != nil {
		return "", fmt.Errorf("failed to validate selection. %w", err)
	}
	return selection, nil
}
