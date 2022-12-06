package server

import (
	"fmt"
	"log"
	"net"
)

func Run(port uint, host string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()
	log.Printf("Waiting for clients on port: %d", port)

	var waitingPlayer *Player
	accept := newAccept(listener)
	join := make(chan net.Conn)
	disconnect := make(chan net.Conn)

	conns := make(map[net.Conn]*Player)

	for {
		select {
		case conn := <-accept:
			conns[conn] = NewPlayer(conn)
			log.Printf("Connection %v added (conns: %d)", conn, len(conns))
			go processConnection(disconnect, join, conns[conn])
		case conn := <-join:
			player := conns[conn]
			if waitingPlayer == nil {
				waitingPlayer = player
			} else {
				session := NewSession(player, waitingPlayer)
				waitingPlayer = nil
				if err := session.Start(); err != nil {
					return err
				}
			}
		case conn := <-disconnect:
			if player, found := conns[conn]; found {
				player.Session.Close()
				if waitingPlayer == player {
					waitingPlayer = nil
					log.Printf("Player %#p (%s) left matchmaker.", &player, player.Name)
				}
				log.Printf("Player %q left.", conns[conn].Name)
				delete(conns, conn)
				log.Printf("Connection %v removed (conns: %d)", conn, len(conns))
			}
		}
	}
}

func newAccept(listener net.Listener) <-chan net.Conn {
	accept := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting incoming connection: %v", err.Error())
			} else {
				accept <- conn
			}
		}
	}()
	return accept
}
