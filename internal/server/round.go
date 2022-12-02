package server

import (
	"log"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

// RoundResult represents a game round outcome.
type RoundResult int

const (
	ResultDraw       RoundResult = 0
	ResultPlayer1Win RoundResult = 1
	ResultPlayer2Win RoundResult = 2
)

func runRound(player, opponent *Player) {
	result := ResultDraw
	for result == ResultDraw {
		log.Printf("Starting a new round. Waiting for player selections...")
		selection1 := game.SelectionNone
		selection2 := game.SelectionNone
		for selection1 == game.SelectionNone || selection2 == game.SelectionNone {
			select {
			case selection := <-player.Selection:
				selection1 = selection
			case selection := <-opponent.Selection:
				selection2 = selection
			}
		}
		result = handleResult(selection1, selection2, player, opponent)
	}
	player.Finished <- struct{}{}
	opponent.Finished <- struct{}{}
}

func handleResult(selection1, selection2 game.Selection, player, opponent *Player) RoundResult {
	result := resolveResult(selection1, selection2)
	log.Printf("Session %q and %q result: %d", player.Name, opponent.Name, result)
	var result1 game.Result
	var result2 game.Result
	switch result {
	case ResultDraw:
		result1 = game.ResultDraw
		result2 = game.ResultDraw
	case ResultPlayer1Win:
		result1 = game.ResultWin
		result2 = game.ResultLose
	case ResultPlayer2Win:
		result1 = game.ResultLose
		result2 = game.ResultWin
	default:
		panic("invalid result!")
	}
	messageContent := com.ResultContent{OpponentSelection: selection2, Result: result1}
	if err := com.WriteMessage(player.Conn, com.TypeResult, messageContent); err != nil {
		log.Fatalf("Failed to send result. %s", err)
	}
	messageContent = com.ResultContent{OpponentSelection: selection1, Result: result2}
	if err := com.WriteMessage(opponent.Conn, com.TypeResult, messageContent); err != nil {
		log.Fatalf("Failed to send result. %s", err)
	}
	return result
}

func resolveResult(player1, player2 game.Selection) RoundResult {
	switch {
	case player1 == player2:
		return ResultDraw
	case player1 == game.SelectionPaper && player2 == game.SelectionScissors:
		return ResultPlayer2Win
	case player1 == game.SelectionRock && player2 == game.SelectionPaper:
		return ResultPlayer2Win
	case player1 == game.SelectionScissors && player2 == game.SelectionRock:
		return ResultPlayer2Win
	default:
		return ResultPlayer1Win
	}
}
