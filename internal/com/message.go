package com

import "encoding/json"

// MessageType specifies the type of the network message.
type MessageType string

const (
	MessageTypeConnect   MessageType = "CONNECT"   // Connection is being opened.
	MessageTypeConnected MessageType = "CONNECTED" // Connection was opened.
)

// Message is base structure for each message being sent between the nodes.
type Message struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content,omitempty"`
}

// ConnectContent contains the content payload for the CONNECT message.
type ConnectContent struct {
	Name string `json:"name"`
}
