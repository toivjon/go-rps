package game_test

import (
	"errors"
	"testing"

	"github.com/toivjon/go-rps/internal/game"
)

func TestSelectionBeats(t *testing.T) {
	t.Parallel()
	t.Run("PanicsWhenLhsSelectionIsNone", func(t *testing.T) {
		defer func() { recover() }()
		game.SelectionNone.Beats(game.SelectionRock)
		t.Fatalf("Expected to panic, but did not!")
	})
	t.Run("PanicsWhenRhsSelectionIsNone", func(t *testing.T) {
		defer func() { recover() }()
		game.SelectionRock.Beats(game.SelectionNone)
		t.Fatalf("Expected to panic, but did not!")
	})
	t.Run("ReturnsTrueWhenWinning", func(t *testing.T) {
		if !game.SelectionRock.Beats(game.SelectionScissors) {
			t.Fatalf("Expected rock to beat scissors, but did not!")
		}
		if !game.SelectionPaper.Beats(game.SelectionRock) {
			t.Fatalf("Expected paper to beat rock, but did not!")
		}
		if !game.SelectionScissors.Beats(game.SelectionPaper) {
			t.Fatalf("Expected scissors to beat paper, but did not!")
		}
	})
	t.Run("ReturnsFalseWhenOtherWins", func(t *testing.T) {
		if game.SelectionRock.Beats(game.SelectionPaper) {
			t.Fatalf("Expected rock to not beat paper, but it did!")
		}
		if game.SelectionPaper.Beats(game.SelectionScissors) {
			t.Fatalf("Expected paper to not beat scissors, but it did!")
		}
		if game.SelectionScissors.Beats(game.SelectionRock) {
			t.Fatalf("Expected scissors to not beat rock, but it did!")
		}
	})
	t.Run("ReturnsFalseWhenDraw", func(t *testing.T) {
		if game.SelectionRock.Beats(game.SelectionRock) {
			t.Fatalf("Expected rock to not beat rock, but it did!")
		}
		if game.SelectionPaper.Beats(game.SelectionPaper) {
			t.Fatalf("Expected paper to not beat paper, but it did!")
		}
		if game.SelectionScissors.Beats(game.SelectionScissors) {
			t.Fatalf("Expected scissors to not beat scissors, but it did!")
		}
	})
}

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
