package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

	matchmaker := make(map[net.Conn]*server.Player)

	for {
		select {
		case conn := <-accept:
			conns[conn] = &server.Player{Conn: conn, Name: "", Selection: make(chan string), Finished: make(chan struct{})}
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

func processConnection(conn net.Conn, disconnect chan<- net.Conn, player *server.Player, join chan<- *server.Player) {
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

	reader := func() chan string {
		outbox := make(chan string)
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

func sendStart(writer io.Writer, opponentName string) error {
	content, err := json.Marshal(com.StartContent{OpponentName: opponentName})
	if err != nil {
		return fmt.Errorf("failed to marshal START content into JSON. %w", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeStart, Content: content}); err != nil {
		return fmt.Errorf("failed to write START message to connection. %w", err)
	}
	return nil
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

func sendResult(writer io.Writer, opponentSelection string, result string) error {
	content, err := json.Marshal(com.ResultContent{OpponentSelection: opponentSelection, Result: result})
	if err != nil {
		return fmt.Errorf("failed to marshal RESULT content into JSON. %w", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeResult, Content: content}); err != nil {
		return fmt.Errorf("failed to write RESULT message to connection. %w", err)
	}
	return nil
}

func enterMatchmaker(matchmaker map[net.Conn]*server.Player, player *server.Player) {
	if len(matchmaker) > 0 {
		runSession(matchmaker, player)
	} else {
		matchmaker[player.Conn] = player
		log.Printf("Player %q joined matchmaker.", player.Name)
	}
}

func runSession(matchmaker map[net.Conn]*server.Player, player *server.Player) {
	for opponentConn, opponent := range matchmaker {
		log.Printf("Starting session %q vs %q", player.Name, opponent.Name)
		if err := sendStart(player.Conn, opponent.Name); err != nil {
			log.Panicln(err)
		}
		if err := sendStart(opponentConn, player.Name); err != nil {
			log.Panicln(err)
		}
		go runRound(player, opponent)
		log.Printf("Sent start message to %v and %v", player.Conn, opponent.Conn)
		return
	}
}

func runRound(player, opponent *server.Player) {
	result := "DRAW"
	for result == "DRAW" {
		log.Printf("Starting a new round. Waiting for player selections...")
		selection1 := ""
		selection2 := ""
		for selection1 == "" || selection2 == "" {
			select {
			case selection := <-player.Selection:
				selection1 = selection
			case selection := <-opponent.Selection:
				selection2 = selection
			}
		}
		// ... resolve results and assign to result variable.
		log.Printf("Session %q and %q result: %s", player.Name, opponent.Name, result)
		if err := sendResult(player.Conn, selection2, result); err != nil {
			log.Fatalf("Failed to send result. %s", err)
		}
		if err := sendResult(opponent.Conn, selection1, result); err != nil {
			log.Fatalf("Failed to send result. %s", err)
		}
	}
	player.Finished <- struct{}{}
	opponent.Finished <- struct{}{}
}

func leaveMatchmaker(matchmaker map[net.Conn]*server.Player, player *server.Player) {
	delete(matchmaker, player.Conn)
	log.Printf("Player %q left matchmaker.", player.Name)
}
