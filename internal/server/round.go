package server

import (
	"github.com/toivjon/go-rps/internal/game"
)

// Round represents a single game round in a RPS game.
type Round struct {
	Selection1 game.Selection
	Selection2 game.Selection
}

// NewRound builds a new round with empty selections.
func NewRound() *Round {
	return &Round{
		Selection1: game.SelectionNone,
		Selection2: game.SelectionNone,
	}
}

// Ended checks whether the round has been ended i.e. both selections have been made.
func (r *Round) Ended() bool {
	return r.Selection1 != game.SelectionNone && r.Selection2 != game.SelectionNone
}

// Result returns game result based on the current selections.
func (r *Round) Result() (game.Result, game.Result) {
	switch {
	case r.Selection1 == r.Selection2:
		return game.ResultDraw, game.ResultDraw
	case r.Selection1.Beats(r.Selection2):
		return game.ResultWin, game.ResultLose
	default:
		return game.ResultLose, game.ResultWin
	}
}
