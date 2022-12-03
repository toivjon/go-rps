package client

import (
	"errors"
	"fmt"
	"log"
	"net"
)

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
