package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
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

	matchmaker := make(map[net.Conn]*Player)

	for {
		select {
		case conn := <-accept:
			conns[conn] = &Player{
				Conn:      conn,
				Name:      "",
				Selection: make(chan game.Selection),
				Finished:  make(chan struct{}),
			}
			log.Printf("Connection %v added (conns: %d)", conn, len(conns))
			go processConnection(conn, disconnect, conns[conn], join)
		case player := <-join:
			conns[player.Conn].Name = player.Name
			log.Printf("Player %q joined.", player.Name)
			enterMatchmaker(matchmaker, player)
		case conn := <-disconnect:
			if player, found := conns[conn]; found {
				leaveMatchmaker(matchmaker, player)
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

func processConnection(conn net.Conn, disconnect chan<- net.Conn, player *Player, join chan<- *Player) {
	defer func() {
		disconnect <- conn
		conn.Close()
	}()

	joinContent, err := readJoin(conn)
	if err != nil {
		disconnect <- conn
		return
	}
	player.Name = joinContent.Name
	join <- player

	reader := func() chan game.Selection {
		outbox := make(chan game.Selection)
		go func() {
			for {
				selection, err := readSelect(conn)
				if err != nil {
					disconnect <- conn
					return
				}
				outbox <- selection.Selection
			}
		}()
		return outbox
	}()

	for {
		log.Printf("Waiting for player %q selection or finished signal...", player.Name)
		select {
		case selection := <-reader:
			log.Printf("Received player %q selection: %s", player.Name, selection)
			player.Selection <- selection
		case <-player.Finished:
			return
		}
	}
}

func readJoin(conn net.Conn) (com.JoinContent, error) {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		return com.JoinContent{}, fmt.Errorf("failed to read JOIN message. %w", err)
	}
	content := com.JoinContent{Name: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.JoinContent{}, fmt.Errorf("failed to read JOIN content. %w", err)
	}
	return content, nil
}

func readSelect(conn net.Conn) (com.SelectContent, error) {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		return com.SelectContent{}, fmt.Errorf("failed to read SELECT message. %w", err)
	}
	content := com.SelectContent{Selection: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.SelectContent{}, fmt.Errorf("failed to read SELECT content. %w", err)
	}
	return content, nil
}

func enterMatchmaker(matchmaker map[net.Conn]*Player, player *Player) {
	if len(matchmaker) > 0 {
		runSession(matchmaker, player)
	} else {
		matchmaker[player.Conn] = player
		log.Printf("Player %q joined matchmaker.", player.Name)
	}
}

func runSession(matchmaker map[net.Conn]*Player, player *Player) {
	for opponentConn, opponent := range matchmaker {
		delete(matchmaker, opponent.Conn)
		log.Printf("Starting session %q vs %q", player.Name, opponent.Name)
		if err := com.WriteMessage(player.Conn, com.TypeStart, com.StartContent{OpponentName: opponent.Name}); err != nil {
			log.Panicln(err)
		}
		if err := com.WriteMessage(opponentConn, com.TypeStart, com.StartContent{OpponentName: player.Name}); err != nil {
			log.Panicln(err)
		}
		go runRound(player, opponent)
		log.Printf("Sent start message to %v and %v", player.Conn, opponent.Conn)
		return
	}
}

func leaveMatchmaker(matchmaker map[net.Conn]*Player, player *Player) {
	delete(matchmaker, player.Conn)
	log.Printf("Player %q left matchmaker.", player.Name)
}
