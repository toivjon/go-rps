package game_test

import (
	"errors"
	"testing"

	"github.com/toivjon/go-rps/internal/game"
)

type match struct {
	s1 game.Selection
	s2 game.Selection
}

func TestSelectionBeats(t *testing.T) {
	t.Parallel()
	t.Run("PanicsWhenLhsSelectionIsNone", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Recover did not have a non-nil error!")
			}
		}()
		game.SelectionNone.Beats(game.SelectionRock)
		t.Fatalf("Expected to panic, but did not!")
	})
	t.Run("PanicsWhenRhsSelectionIsNone", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Recover did not have a non-nil error!")
			}
		}()
		game.SelectionRock.Beats(game.SelectionNone)
		t.Fatalf("Expected to panic, but did not!")
	})
	t.Run("ReturnsTrueWhenWinning", func(t *testing.T) {
		t.Parallel()
		matches := []match{
			{s1: game.SelectionRock, s2: game.SelectionScissors},
			{s1: game.SelectionPaper, s2: game.SelectionRock},
			{s1: game.SelectionScissors, s2: game.SelectionPaper},
		}
		for _, match := range matches {
			if !match.s1.Beats(match.s2) {
				t.Fatalf("Expected %q to beat %q, but it did not!", match.s1, match.s2)
			}
		}
	})
	t.Run("ReturnsFalseWhenOtherWins", func(t *testing.T) {
		t.Parallel()
		matches := []match{
			{s1: game.SelectionRock, s2: game.SelectionPaper},
			{s1: game.SelectionPaper, s2: game.SelectionScissors},
			{s1: game.SelectionScissors, s2: game.SelectionRock},
		}
		for _, match := range matches {
			if match.s1.Beats(match.s2) {
				t.Fatalf("Expected %q to not beat %q, but it did!", match.s1, match.s2)
			}
		}
	})
	t.Run("ReturnsFalseWhenDraw", func(t *testing.T) {
		t.Parallel()
		selections := []game.Selection{game.SelectionRock, game.SelectionPaper, game.SelectionScissors}
		for _, selection := range selections {
			if selection.Beats(selection) {
				t.Fatalf("Expected %q to not beat %q, but it did!", selection, selection)
			}
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
