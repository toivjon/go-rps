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

// ErrInvalidSelection is an error occurring when selection validation fails.
var ErrInvalidSelection = errors.New("the provided value contains an invalid selection")

// ValidateSelection returns an error if the given selection is not valid..
func ValidateSelection(val Selection) error {
	if val != SelectionRock && val != SelectionPaper && val != SelectionScissors {
		return ErrInvalidSelection
	}
	return nil
}
