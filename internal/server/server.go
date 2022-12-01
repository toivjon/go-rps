package server

import (
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/game"
)

func Run(port uint, host string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()
	log.Printf("Waiting for clients on port: %d", port)

	accept := newAccept(listener)
	disconnect := make(chan net.Conn)
	join := make(chan *Player)

	conns := make(map[net.Conn]*Player)

	matchmaker := new(matchmaker)
	for {
		select {
		case conn := <-accept:
			conns[conn] = &Player{
				Conn:      conn,
				Name:      "",
				Selection: make(chan game.Selection),
				Finished:  make(chan struct{}),
				Session:   nil,
			}
			log.Printf("Connection %v added (conns: %d)", conn, len(conns))
			go processConnection(conn, disconnect, conns[conn], join)
		case player := <-join:
			conns[player.Conn].Name = player.Name
			if err := matchmaker.Enter(player); err != nil {
				log.Printf("Player %q Failed to enter matchmaker. %s", player.Name, err)
			} else {
				log.Printf("Player %q joined.", player.Name)
			}
		case conn := <-disconnect:
			if player, found := conns[conn]; found {
				player.Session.Close()
				matchmaker.Leave(player)
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
