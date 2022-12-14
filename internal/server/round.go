package server

import (
	"github.com/toivjon/go-rps/internal/game"
)

// Round represents a single game round in a RPS game.
type Round struct {
	selection1 game.Selection
	selection2 game.Selection
}

// NewRound builds a new round with empty selections.
func NewRound() *Round {
	return &Round{
		selection1: game.SelectionNone,
		selection2: game.SelectionNone,
	}
}

// Ended checks whether the round has been ended i.e. both selections have been made.
func (r *Round) Ended() bool {
	return r.selection1 != game.SelectionNone && r.selection2 != game.SelectionNone
}

// Result returns game result based on the current selections.
func (r *Round) Result() (game.Result, game.Result) {
	switch {
	case r.selection1 == r.selection2:
		return game.ResultDraw, game.ResultDraw
	case r.selection1.Beats(r.selection2):
		return game.ResultWin, game.ResultLose
	default:
		return game.ResultLose, game.ResultWin
	}
}
