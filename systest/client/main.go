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
	joinContent := mustRead[com.JoinContent](conn, com.TypeJoin)
	assertName("anonymous", joinContent.Name)
	mustSend(conn, com.TypeStart, com.StartContent{OpponentName: "mickey"})
	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection := mustRead[com.SelectContent](conn, com.TypeSelect)
	assertSelection(game.SelectionRock, selection.Selection)
	mustSend(conn, com.TypeResult, com.ResultContent{OpponentSelection: game.SelectionPaper, Result: game.ResultLose})
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
	joinContent := mustRead[com.JoinContent](conn, com.TypeJoin)
	assertName("anonymous", joinContent.Name)
	mustSend(conn, com.TypeStart, com.StartContent{OpponentName: "mickey"})

	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection := mustRead[com.SelectContent](conn, com.TypeSelect)
	assertSelection(game.SelectionRock, selection.Selection)
	mustSend(conn, com.TypeResult, com.ResultContent{OpponentSelection: game.SelectionRock, Result: game.ResultDraw})

	if _, err := input.Write([]byte("p\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = mustRead[com.SelectContent](conn, com.TypeSelect)
	assertSelection(game.SelectionPaper, selection.Selection)
	mustSend(conn, com.TypeResult, com.ResultContent{OpponentSelection: game.SelectionPaper, Result: game.ResultDraw})

	if _, err := input.Write([]byte("s\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = mustRead[com.SelectContent](conn, com.TypeSelect)
	assertSelection(game.SelectionScissors, selection.Selection)
	mustSend(conn, com.TypeResult, com.ResultContent{OpponentSelection: game.SelectionScissors, Result: game.ResultDraw})

	if _, err := input.Write([]byte("r\n")); err != nil {
		log.Panicf("Failed to write data to client stdin. %s", err)
	}
	selection = mustRead[com.SelectContent](conn, com.TypeSelect)
	assertSelection(game.SelectionRock, selection.Selection)
	mustSend(conn, com.TypeResult, com.ResultContent{OpponentSelection: game.SelectionScissors, Result: game.ResultWin})

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

func mustRead[T any](reader io.Reader, messageType com.MessageType) *T {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		log.Panicf("Failed to read %s message. %s", messageType, err)
	}
	if message.Type != messageType {
		log.Panicf("Unexpected message type received. Expected: %s Was: %s", messageType, message.Type)
	}
	content := new(T)
	if err := json.Unmarshal(message.Content, content); err != nil {
		log.Panicf("Failed to read %s content. %s", messageType, err)
	}
	return content
}

func mustSend[T any](writer io.Writer, messageType com.MessageType, val T) {
	content, err := json.Marshal(val)
	if err != nil {
		log.Panicf("Failed to marshal %s content into JSON. %s", messageType, err)
	}
	message := com.Message{Type: messageType, Content: content}
	if err := com.Write(writer, message); err != nil {
		log.Panicf("Failed to write %s message to connection. %s", messageType, err)
	}
}
