package server

import "log"

type matchmaker struct {
	waitingPlayer *Player
}

func (m *matchmaker) Enter(player *Player) error {
	if m.waitingPlayer == nil {
		m.waitingPlayer = player
		log.Printf("Player %#p (%s) entered matchmaker.", player, player.Name)
		return nil
	}
	session := newSession(player, m.waitingPlayer)
	m.waitingPlayer = nil
	if err := session.Start(); err != nil {
		return err
	}
	return nil
}

func (m *matchmaker) Leave(player *Player) {
	if m.waitingPlayer == player {
		m.waitingPlayer = nil
		log.Printf("Player %#p (%s) left matchmaker.", &player, player.Name)
	}
}
