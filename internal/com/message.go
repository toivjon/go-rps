package com

// ConnectMessage is the initial client-to-server message which also contains the player name.
type ConnectMessage struct {
	Name string `json:"name"`
}

// ConnectedMessage is the response to client's initial message.
type ConnectedMessage struct {
	OK bool `json:"ok"`
}
