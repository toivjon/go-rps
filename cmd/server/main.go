package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/server"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

func main() {
	port := flag.Uint("port", defaultPort, "The port to listen for connections.")
	host := flag.String("host", defaultHost, "The network address to listen for connections.")
	flag.Parse()

	log.Println("Starting RPS server...")
	if err := start(*port, *host); err != nil {
		log.Fatal(err)
	}
}

func start(port uint, host string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()
	log.Printf("Waiting for clients on port: %d", port)

	accept := newAccept(listener)
	disconnect := make(chan net.Conn)
	join := make(chan *server.Player)

	conns := make(map[net.Conn]*server.Player)

	matchmaker := make(map[net.Conn]bool)

	for {
		select {
		case conn := <-accept:
			conns[conn] = &server.Player{Conn: conn, Name: ""}
			log.Printf("Connection %v added (conns: %d)", conn, len(conns))
			go processConnection(conn, disconnect, conns[conn], join)
		case player := <-join:
			conns[player.Conn].Name = player.Name
			log.Printf("Player %q joined.", player.Name)
			enterMatchmaker(matchmaker, *player)
		case conn := <-disconnect:
			if player, found := conns[conn]; found {
				leaveMatchmaker(matchmaker, *player)
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

func processConnection(conn net.Conn, disconnect chan<- net.Conn, player *server.Player, join chan<- *server.Player) {
	defer func() {
		disconnect <- conn
		conn.Close()
	}()

	input, err := com.Read[com.Message](conn)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
		disconnect <- conn
		return
	}

	content := new(com.ConnectContent)
	if err := json.Unmarshal(input.Content, content); err != nil {
		log.Printf("Failed to unmarshal message content. %v", err)
		disconnect <- conn
	}

	player.Name = content.Name

	log.Printf("Read message: %+v content: %+v", input, content)

	if err := com.WriteConnected(conn); err != nil {
		log.Printf("Failed to write data: %s", err)
		return
	}

	join <- player

	for {
		_, err = com.Read[com.Message](conn)
		if err != nil {
			log.Printf("Failed to read data: %s", err)
			disconnect <- conn
			return
		}
	}
}

func enterMatchmaker(matchmaker map[net.Conn]bool, player server.Player) {
	if len(matchmaker) > 0 {
		// ... match found! start a game session.
		log.Printf("Matchmaker found an opponent. Let the game begin!")
	} else {
		matchmaker[player.Conn] = true
		log.Printf("Player %q joined matchmaker.", player.Name)
	}
}

func leaveMatchmaker(matchmaker map[net.Conn]bool, player server.Player) {
	delete(matchmaker, player.Conn)
	log.Printf("Player %q left matchmaker.", player.Name)
}
