package server_test

import (
	"testing"

	"github.com/toivjon/go-rps/internal/game"
	"github.com/toivjon/go-rps/internal/server"
)

func TestNewRound(t *testing.T) {
	t.Parallel()
	round := server.NewRound()
	if round.Selection1 != game.SelectionNone {
		t.Fatalf("Expected selection one to be %q, but was %q!", game.SelectionNone, round.Selection1)
	}
	if round.Selection2 != game.SelectionNone {
		t.Fatalf("Expected selection two to be %q, but was %q!", game.SelectionNone, round.Selection2)
	}
}

func TestRoundEnded(t *testing.T) {
	t.Parallel()
	t.Run("ReturnFalseWhenSelection1IsNone", func(t *testing.T) {
		t.Parallel()
		round := server.Round{Selection1: game.SelectionNone, Selection2: game.SelectionRock}
		if round.Ended() {
			t.Fatal("Expected to return false when Selection1 is none, but returned true!")
		}
	})
	t.Run("ReturnFalseWhenSelection2IsNone", func(t *testing.T) {
		t.Parallel()
		round := server.Round{Selection1: game.SelectionRock, Selection2: game.SelectionNone}
		if round.Ended() {
			t.Fatal("Expected to return false when Selection2 is none, but returned true!")
		}
	})
	t.Run("ReturnTrueWhenBothSelectionsAreNotNone", func(t *testing.T) {
		t.Parallel()
		round := server.Round{Selection1: game.SelectionRock, Selection2: game.SelectionRock}
		if !round.Ended() {
			t.Fatal("Expected to return true both selections are not none, but returned false!")
		}
	})
}

func TestRoundResult(t *testing.T) {
	t.Parallel()
	t.Run("ReturnDrawsWhenSelectionsAreSame", func(t *testing.T) {
		t.Parallel()
		for _, selection := range []game.Selection{game.SelectionPaper, game.SelectionRock, game.SelectionScissors} {
			round := server.Round{Selection1: selection, Selection2: selection}
			result1, result2 := round.Result()
			if result1 != game.ResultDraw {
				t.Fatalf("Expected result1 to be %q, but was %q!", game.ResultDraw, result1)
			}
			if result2 != game.ResultDraw {
				t.Fatalf("Expected result2 to be %q, but was %q!", game.ResultDraw, result2)
			}
		}
	})
	t.Run("ReturnWinAndLoseWhenSelection1Wins", func(t *testing.T) {
		t.Parallel()
		roundSelections := [][]game.Selection{
			{game.SelectionRock, game.SelectionScissors},
			{game.SelectionPaper, game.SelectionRock},
			{game.SelectionScissors, game.SelectionPaper},
		}
		for _, selections := range roundSelections {
			round := server.Round{Selection1: selections[0], Selection2: selections[1]}
			result1, result2 := round.Result()
			if result1 != game.ResultWin {
				t.Fatalf("Expected result1 to be %q, but was %q!", game.ResultWin, result1)
			}
			if result2 != game.ResultLose {
				t.Fatalf("Expected result2 to be %q, but was %q!", game.ResultLose, result2)
			}
		}
	})
	t.Run("ReturnLoseAndWinWhenSelection2Wins", func(t *testing.T) {
		t.Parallel()
		roundSelections := [][]game.Selection{
			{game.SelectionRock, game.SelectionPaper},
			{game.SelectionPaper, game.SelectionScissors},
			{game.SelectionScissors, game.SelectionRock},
		}
		for _, selections := range roundSelections {
			round := server.Round{Selection1: selections[0], Selection2: selections[1]}
			result1, result2 := round.Result()
			if result1 != game.ResultLose {
				t.Fatalf("Expected result1 to be %q, but was %q!", game.ResultLose, result1)
			}
			if result2 != game.ResultWin {
				t.Fatalf("Expected result2 to be %q, but was %q!", game.ResultWin, result2)
			}
		}
	})
}
