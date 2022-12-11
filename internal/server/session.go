package server

import "github.com/toivjon/go-rps/internal/game"

// Session represents a single game session where to clients battle against each other in RPS rounds.
type Session struct {
	cli1          *Client
	cli2          *Client
	cli1Selection game.Selection
	cli2Selection game.Selection
}

// NewSession builds a new session for  the given clients.
func NewSession(cli1, cli2 *Client) *Session {
	return &Session{
		cli1:          cli1,
		cli2:          cli2,
		cli1Selection: game.SelectionNone,
		cli2Selection: game.SelectionNone,
	}
}
