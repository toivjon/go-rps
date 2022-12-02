package server

import (
	"fmt"
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

type Round struct {
	session     *Session
	p1Selection game.Selection
	p2Selection game.Selection
}

func NewRound(session *Session) *Round {
	return &Round{
		session:     session,
		p1Selection: game.SelectionNone,
		p2Selection: game.SelectionNone,
	}
}

func (r *Round) Play() (RoundResult, error) {
	log.Printf("Starting a new round. Waiting for player selections...")
	for r.p1Selection == game.SelectionNone || r.p2Selection == game.SelectionNone {
		select {
		case r.p1Selection = <-r.session.player1.Selection:
		case r.p2Selection = <-r.session.player2.Selection:
		}
	}
	result := r.resolveResult()
	if err := r.handleResult(result); err != nil {
		return result, fmt.Errorf("failed to handle round result. %w", err)
	}
	return result, nil
}

func (r *Round) resolveResult() RoundResult {
	switch {
	case r.p1Selection == r.p2Selection:
		return ResultDraw
	case r.p1Selection == game.SelectionPaper && r.p2Selection == game.SelectionScissors:
		return ResultPlayer2Win
	case r.p1Selection == game.SelectionRock && r.p2Selection == game.SelectionPaper:
		return ResultPlayer2Win
	case r.p1Selection == game.SelectionScissors && r.p2Selection == game.SelectionRock:
		return ResultPlayer2Win
	default:
		return ResultPlayer1Win
	}
}

func (r *Round) handleResult(result RoundResult) error {
	log.Printf("Session %q and %q result: %d", r.session.player1.Name, r.session.player2.Name, result)
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
	messageContent := com.ResultContent{OpponentSelection: r.p2Selection, Result: result1}
	if err := com.WriteMessage(r.session.player1.Conn, com.TypeResult, messageContent); err != nil {
		return fmt.Errorf("failed to write result message for client %#p. %w", r.session.player1, err)
	}
	messageContent = com.ResultContent{OpponentSelection: r.p1Selection, Result: result2}
	if err := com.WriteMessage(r.session.player2.Conn, com.TypeResult, messageContent); err != nil {
		return fmt.Errorf("failed to write result message for client %#p. %w", r.session.player2, err)
	}
	return nil
}
