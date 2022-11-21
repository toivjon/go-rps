package com

import (
	"encoding/json"

	"github.com/toivjon/go-rps/internal/game"
)

// MessageType specifies the type of the network message.
type MessageType string

const (
	TypeJoin   MessageType = "JOIN"   // Client wants to join server.
	TypeStart  MessageType = "START"  // Server starts a game session.
	TypeSelect MessageType = "SELECT" // Client decides an in-game decision.
	TypeResult MessageType = "RESULT" // Server resolves game session round result.
)

// Message is base structure for each message being sent between the nodes.
type Message struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content"`
}

// JoinContent contains the content of a JOIN message.
type JoinContent struct {
	Name string
}

// StartContent contains the content of a START message.
type StartContent struct {
	OpponentName string
}

// SelectContent contains the content of a SELECT message.
type SelectContent struct {
	Selection game.Selection
}

// ResultContent contains the content of a RESULT message.
type ResultContent struct {
	OpponentSelection game.Selection
	Result            game.Result
}
