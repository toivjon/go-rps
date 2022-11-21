package game

// Result represents a game session round outcome.
type Result string

const (
	ResultWin  Result = "WIN"
	ResultLose Result = "LOSE"
	ResultDraw Result = "DRAW"
)
