package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

func Start(port uint, host string) error {
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

const (
	ResultDraw       = 0
	ResultPlayer1Win = 1
	ResultPlayer2Win = 2
)

func runRound(player, opponent *Player) {
	result := ResultDraw
	for result == ResultDraw {
		log.Printf("Starting a new round. Waiting for player selections...")
		selection1 := game.SelectionNone
		selection2 := game.SelectionNone
		for selection1 == game.SelectionNone || selection2 == game.SelectionNone {
			select {
			case selection := <-player.Selection:
				selection1 = selection
			case selection := <-opponent.Selection:
				selection2 = selection
			}
		}
		result = handleResult(selection1, selection2, player, opponent)
	}
	player.Finished <- struct{}{}
	opponent.Finished <- struct{}{}
}

func leaveMatchmaker(matchmaker map[net.Conn]*Player, player *Player) {
	delete(matchmaker, player.Conn)
	log.Printf("Player %q left matchmaker.", player.Name)
}

func resolveResult(player1, player2 game.Selection) int {
	switch {
	case player1 == player2:
		return ResultDraw
	case player1 == game.SelectionPaper && player2 == game.SelectionScissors:
		return ResultPlayer2Win
	case player1 == game.SelectionRock && player2 == game.SelectionPaper:
		return ResultPlayer2Win
	case player1 == game.SelectionScissors && player2 == game.SelectionRock:
		return ResultPlayer2Win
	default:
		return ResultPlayer1Win
	}
}

func handleResult(selection1, selection2 game.Selection, player, opponent *Player) int {
	result := resolveResult(selection1, selection2)
	log.Printf("Session %q and %q result: %d", player.Name, opponent.Name, result)
	var result1 game.Result
	var result2 game.Result
	switch result {
	case ResultDraw:
		result1 = game.ResultDraw
		result2 = game.ResultDraw
	case ResultPlayer1Win:
		result1 = game.ResultWin
		result2 = game.ResultLose
	case ResultPlayer2Win:
		result1 = game.ResultLose
		result2 = game.ResultWin
	default:
		panic("invalid result!")
	}
	messageContent := com.ResultContent{OpponentSelection: selection2, Result: result1}
	if err := com.WriteMessage(player.Conn, com.TypeResult, messageContent); err != nil {
		log.Fatalf("Failed to send result. %s", err)
	}
	messageContent = com.ResultContent{OpponentSelection: selection1, Result: result2}
	if err := com.WriteMessage(opponent.Conn, com.TypeResult, messageContent); err != nil {
		log.Fatalf("Failed to send result. %s", err)
	}
	return result
}
