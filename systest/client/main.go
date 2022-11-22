package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

const (
	serverPort = 7777
	serverHost = "localhost"
)

func main() {
	log.SetPrefix("[system-test] ")
	log.Println("Running client system tests...")
	testReturnErrorWhenConnectingFails()
	testPlaySessionWithOneRound()
	testPlaySessionWithManyRounds()
}

func testReturnErrorWhenConnectingFails() {
	log.Println("Test that client exits with error code (1) if initial connection fails.")
	client, _ := startClient()
	state, err := client.Process.Wait()

	assertNoError(err)
	assertExited(state)
	assertExitCode(1, state.ExitCode())
}

func testPlaySessionWithOneRound() {
	log.Println("Test that client logic works in a session with one round.")
	server := startServer()
	defer closeServer(server)

	client, input := startClient()
	defer closeClient(client)

	conn := accept(server)
	joinContent := readJoin(conn)
	assertName("anonymous", joinContent.Name)
	sendStart(conn, "mickey")
	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection := readSelect(conn)
	assertSelection(game.SelectionRock, selection.Selection)
	sendResult(conn, game.SelectionPaper, game.ResultLose)
	if err := client.Wait(); err != nil {
		log.Panicf("Unable to wait until client disconnects and closes. %s", err)
	}
}

func testPlaySessionWithManyRounds() {
	log.Println("Test that client logic works in a session with many rounds.")
	server := startServer()
	defer closeServer(server)

	client, input := startClient()
	defer closeClient(client)

	conn := accept(server)
	joinContent := readJoin(conn)
	assertName("anonymous", joinContent.Name)
	sendStart(conn, "mickey")

	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection := readSelect(conn)
	assertSelection(game.SelectionRock, selection.Selection)
	sendResult(conn, game.SelectionRock, game.ResultDraw)

	if _, err := input.Write([]byte("p\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = readSelect(conn)
	assertSelection(game.SelectionPaper, selection.Selection)
	sendResult(conn, game.SelectionPaper, game.ResultDraw)

	if _, err := input.Write([]byte("s\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = readSelect(conn)
	assertSelection(game.SelectionScissors, selection.Selection)
	sendResult(conn, game.SelectionScissors, game.ResultDraw)

	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = readSelect(conn)
	assertSelection(game.SelectionRock, selection.Selection)
	sendResult(conn, game.SelectionScissors, game.ResultWin)

	if err := client.Wait(); err != nil {
		log.Panicf("Unable to wait until client disconnects and closes. %s", err)
	}
}

func startClient() (*exec.Cmd, io.WriteCloser) {
	cmd := exec.Command("./bin/client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	writer, err := cmd.StdinPipe()
	if err != nil {
		log.Panicf("Failed to open stdin pipe. %s", err)
	}
	if err := cmd.Start(); err != nil {
		log.Panicf("Failed to start client process. %s", err)
	}
	return cmd, writer
}

func closeClient(client *exec.Cmd) {
	if !client.ProcessState.Exited() {
		if err := client.Process.Kill(); err != nil {
			log.Panicf("Failed to kill client process. %s", err)
		}
	}
}

func startServer() net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		log.Panicf("Failed to start TCP listener on port %d. %s", serverPort, err)
	}
	return listener
}

func closeServer(listener net.Listener) {
	if err := listener.Close(); err != nil {
		log.Panicf("Failed to close server listener. %s", err)
	}
}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()
	if err != nil {
		log.Panicf("Failed to accept incoming connection. %s", err)
	}
	return conn
}

func assertNoError(err error) {
	if err != nil {
		log.Panicf("Expected no error, but was %q", err)
	}
}

func assertExited(state *os.ProcessState) {
	if !state.Exited() {
		log.Panicf("Expected process state to be exited, but it was not.")
	}
}

func assertExitCode(expected, actual int) {
	if expected != actual {
		log.Panicf("Unexpected exit code. Expected: %d Was: %d", expected, actual)
	}
}

func assertName(expected, actual string) {
	if expected != actual {
		log.Panicf("Unexpected name. Expected: %s Was: %s", expected, actual)
	}
}

func assertSelection(expected, actual game.Selection) {
	if expected != actual {
		log.Panicf("Unexpected selection. Expected: %s Was: %s", expected, actual)
	}
}

func readJoin(conn net.Conn) com.JoinContent {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		log.Panicf("Failed to read JOIN message. %s", err)
	}
	content := com.JoinContent{Name: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		log.Panicf("Failed to read JOIN content. %s", err)
	}
	return content
}

func sendStart(writer io.Writer, opponentName string) {
	content, err := json.Marshal(com.StartContent{OpponentName: opponentName})
	if err != nil {
		log.Panicf("Failed to marshal START content into JSON. %s", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeStart, Content: content}); err != nil {
		log.Panicf("Failed to write START message to connection. %s", err)
	}
}

func readSelect(conn net.Conn) com.SelectContent {
	message, err := com.Read[com.Message](conn)
	if err != nil {
		log.Panicf("Failed to read SELECT message. %s", err)
	}
	content := com.SelectContent{Selection: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		log.Panicf("Failed to read SELECT content. %s", err)
	}
	return content
}

func sendResult(writer io.Writer, opponentSelection game.Selection, result game.Result) {
	content, err := json.Marshal(com.ResultContent{OpponentSelection: opponentSelection, Result: result})
	if err != nil {
		log.Panicf("Failed to marshal RESULT content into JSON. %s", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeResult, Content: content}); err != nil {
		log.Panicf("Failed to write RESULT message to connection. %s", err)
	}
}
