package client

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

// State represents a reference to a client state which may return next state or an error.
type State func(conn net.Conn) (State, error)

func Run(port uint, host, name string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	log.Printf("Joining as player: %s", name)
	if err := com.WriteMessage(conn, com.TypeJoin, com.JoinContent{Name: name}); err != nil {
		return fmt.Errorf("failed to write JOIN message. %w", err)
	}

	state := Joined
	for state != nil {
		state, err = state(conn)
	}
	if err != nil && !errors.Is(err, ErrEnd) {
		return err
	}
	return nil
}
