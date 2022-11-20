package game_test

import (
	"errors"
	"testing"

	"github.com/toivjon/go-rps/internal/game"
)

func TestValidateSelection(t *testing.T) {
	t.Parallel()
	t.Run("ReturnsErrorWithEmptyInput", func(t *testing.T) {
		t.Parallel()
		if err := game.ValidateSelection(game.Selection("")); !errors.Is(err, game.ErrInvalidSelection) {
			t.Fatalf("Expected invalid selection error, but %v was returned!", err)
		}
	})
	t.Run("ReturnsErrorWithInvalidNonEmptyInput", func(t *testing.T) {
		t.Parallel()
		if err := game.ValidateSelection(game.Selection("a")); !errors.Is(err, game.ErrInvalidSelection) {
			t.Fatalf("Expected invalid selection error, but %v was returned!", err)
		}
	})
	t.Run("ReturnsNilWithValidInputs", func(t *testing.T) {
		t.Parallel()
		inputs := []game.Selection{game.SelectionRock, game.SelectionPaper, game.SelectionScissors}
		for _, input := range inputs {
			if err := game.ValidateSelection(input); err != nil {
				t.Fatalf("Expected %s to return no error, but error was returned: %s", input, err)
			}
		}
	})
}
