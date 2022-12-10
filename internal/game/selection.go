package game

import (
	"errors"
)

// Selection represents a player selection in a game round.
type Selection string

const (
	SelectionNone     Selection = ""
	SelectionRock     Selection = "r"
	SelectionPaper    Selection = "p"
	SelectionScissors Selection = "s"
)

// Beats checks whether the selection beats the other selection. Panics if either selection is none.
func (s Selection) Beats(other Selection) bool {
	if s == SelectionNone || other == SelectionNone {
		panic("Unable to check beat with no selection!")
	}
	switch {
	case s == SelectionRock && other == SelectionScissors:
		return true
	case s == SelectionPaper && other == SelectionRock:
		return true
	case s == SelectionScissors && other == SelectionPaper:
		return true
	}
	return false
}

// ErrInvalidSelection is an error occurring when selection validation fails.
var ErrInvalidSelection = errors.New("the provided value contains an invalid selection")

// ValidateSelection returns an error if the given selection is not valid..
func ValidateSelection(val Selection) error {
	if val != SelectionRock && val != SelectionPaper && val != SelectionScissors {
		return ErrInvalidSelection
	}
	return nil
}
