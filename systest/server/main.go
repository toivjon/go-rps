package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

const (
	serverHost = "localhost"
	serverPort = 7777
	name1      = "donald"
	name2      = "mickey"
)

func main() {
	log.SetPrefix("[system-test] ")
	log.Println("Running server system tests...")
	testPlaySessionWithOneRound()
	testPlaySessionWithManyRounds()
	testPlayManySessionsConcurrently()
}

func testPlaySessionWithOneRound() {
	log.Println("Test Play Session With One Round")
	server := startServer()
	defer closeServer(server)
	time.Sleep(time.Second)

	client1 := newClient()
	defer client1.Close()
	client2 := newClient()
	defer client2.Close()

	sendJoin(client1, name1)
	sendJoin(client2, name2)

	start1 := readStart(client1)
	start2 := readStart(client2)
	assertOpponentName(start1, name2)
	assertOpponentName(start2, name1)

	sendSelect(client1, game.SelectionRock)
	sendSelect(client2, game.SelectionPaper)

	result1 := readResult(client1)
	result2 := readResult(client2)
	assertResult(result1, game.SelectionPaper, "DRAW") // ... Non-DRAW is not yet supported.
	assertResult(result2, game.SelectionRock, "DRAW")  // ... Non-DRAW is not yet supported.
}

func testPlaySessionWithManyRounds() {
	log.Println("Test Play Session With Many Rounds")
	server := startServer()
	defer closeServer(server)
	time.Sleep(time.Second)

	client1 := newClient()
	defer client1.Close()
	client2 := newClient()
	defer client2.Close()

	sendJoin(client1, name1)
	sendJoin(client2, name2)

	start1 := readStart(client1)
	start2 := readStart(client2)
	assertOpponentName(start1, name2)
	assertOpponentName(start2, name1)

	sendSelect(client1, game.SelectionRock)
	sendSelect(client2, game.SelectionRock)
	result1 := readResult(client1)
	result2 := readResult(client2)
	assertResult(result1, game.SelectionRock, "DRAW")
	assertResult(result2, game.SelectionRock, "DRAW")

	sendSelect(client1, game.SelectionPaper)
	sendSelect(client2, game.SelectionPaper)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionPaper, "DRAW")
	assertResult(result2, game.SelectionPaper, "DRAW")

	sendSelect(client1, game.SelectionScissors)
	sendSelect(client2, game.SelectionScissors)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionScissors, "DRAW")
	assertResult(result2, game.SelectionScissors, "DRAW")

	sendSelect(client1, game.SelectionScissors)
	sendSelect(client2, game.SelectionPaper)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionPaper, "DRAW")    // ... Non-DRAW is not yet supported.
	assertResult(result2, game.SelectionScissors, "DRAW") // ... Non-DRAW is not yet supported.
}

func testPlayManySessionsConcurrently() {
	log.Println("Test Play Two Sessions Concurrently")
	server := startServer()
	defer closeServer(server)

	// ... Not yet supported.
}

func assertOpponentName(start com.StartContent, expected string) {
	if start.OpponentName != expected {
		log.Panicf("Invalid opponent name. Expected: %q Was: %q", expected, start.OpponentName)
	}
}

//nolint:unparam // Remove this after the real game logic has been implemented!
func assertResult(result com.ResultContent, expectedOpponentSelection game.Selection, expectedResult string) {
	if result.OpponentSelection != expectedOpponentSelection {
		log.Panicf("Invalid opponent selection. Expected: %q Was: %q",
			expectedOpponentSelection,
			result.OpponentSelection,
		)
	}
	if result.Result != expectedResult {
		log.Panicf("Invalid result. Expected: %q Was: %q",
			expectedResult,
			result.Result,
		)
	}
}

func startServer() *exec.Cmd {
	cmd := exec.Command("./bin/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Panicf("Failed to start server process. %s", err)
	}
	return cmd
}

func closeServer(server *exec.Cmd) {
	if err := server.Process.Kill(); err != nil {
		log.Panicf("Failed to kill server process. %s", err)
	}
}

func newClient() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		log.Panicln("Failed to open TCP connection to server. %w", err)
	}
	return conn
}

func sendJoin(writer io.Writer, name string) {
	content, err := json.Marshal(com.JoinContent{Name: name})
	if err != nil {
		log.Panicf("failed marshal JOIN content into JSON. %s", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeJoin, Content: content}); err != nil {
		log.Panicf("failed to write JOIN message to connection. %s", err)
	}
}

func readStart(reader io.Reader) com.StartContent {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		log.Panicf("failed to read START message. %s", err)
	}
	content := com.StartContent{OpponentName: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		log.Panicf("failed to read START content. %s", err)
	}
	return content
}

func sendSelect(writer io.Writer, selection game.Selection) {
	content, err := json.Marshal(com.SelectContent{Selection: selection})
	if err != nil {
		log.Panicf("failed to marshal SELECT content into JSON. %s", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeSelect, Content: content}); err != nil {
		log.Panicf("failed to write SELECT message to connection. %s", err)
	}
}

func readResult(reader io.Reader) com.ResultContent {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		log.Panicf("failed to read RESULT message. %s", err)
	}
	content := com.ResultContent{OpponentSelection: "", Result: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		log.Panicf("failed to read RESULT content. %s", err)
	}
	return content
}
