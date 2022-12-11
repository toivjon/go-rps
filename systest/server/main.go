package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

const (
	serverHost    = "localhost"
	serverPort    = 7777
	startupDelay  = 2 * time.Second
	serverTimeout = 10 * time.Second
	name1         = "donald"
	name2         = "mickey"
)

func main() {
	log.SetPrefix("[system-test] ")
	log.Println("Running server system tests...")
	testPlaySessionWithOneRound()
	testPlaySessionWithManyRounds()
	testPlayManySessionsConcurrently()
	testSessionEndsWhenClientDisconnects()
}

func testPlaySessionWithOneRound() {
	log.Println("Test Play Session With One Round")
	server, cancel := startServer()
	defer closeServer(server, cancel)
	time.Sleep(startupDelay)

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
	assertResult(result1, game.SelectionPaper, game.ResultLose)
	assertResult(result2, game.SelectionRock, game.ResultWin)
}

func testPlaySessionWithManyRounds() {
	log.Println("Test Play Session With Many Rounds")
	server, cancel := startServer()
	defer closeServer(server, cancel)
	time.Sleep(startupDelay)

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
	assertResult(result1, game.SelectionRock, game.ResultDraw)
	assertResult(result2, game.SelectionRock, game.ResultDraw)

	sendSelect(client1, game.SelectionPaper)
	sendSelect(client2, game.SelectionPaper)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionPaper, game.ResultDraw)
	assertResult(result2, game.SelectionPaper, game.ResultDraw)

	sendSelect(client1, game.SelectionScissors)
	sendSelect(client2, game.SelectionScissors)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionScissors, game.ResultDraw)
	assertResult(result2, game.SelectionScissors, game.ResultDraw)

	sendSelect(client1, game.SelectionScissors)
	sendSelect(client2, game.SelectionPaper)
	result1 = readResult(client1)
	result2 = readResult(client2)
	assertResult(result1, game.SelectionPaper, game.ResultWin)
	assertResult(result2, game.SelectionScissors, game.ResultLose)
}

func testPlayManySessionsConcurrently() {
	log.Println("Test Play Two Sessions Concurrently")
	server, cancel := startServer()
	defer closeServer(server, cancel)
	time.Sleep(startupDelay)

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

	client3 := newClient()
	defer client3.Close()
	client4 := newClient()
	defer client4.Close()

	sendJoin(client3, name1)
	sendJoin(client4, name2)

	start3 := readStart(client3)
	start4 := readStart(client4)
	assertOpponentName(start3, name2)
	assertOpponentName(start4, name1)

	sendSelect(client1, game.SelectionRock)
	sendSelect(client2, game.SelectionPaper)
	sendSelect(client3, game.SelectionRock)

	result1 := readResult(client1)
	result2 := readResult(client2)
	assertResult(result1, game.SelectionPaper, game.ResultLose)
	assertResult(result2, game.SelectionRock, game.ResultWin)

	sendSelect(client4, game.SelectionPaper)

	result3 := readResult(client3)
	result4 := readResult(client4)
	assertResult(result3, game.SelectionPaper, game.ResultLose)
	assertResult(result4, game.SelectionRock, game.ResultWin)
}

func testSessionEndsWhenClientDisconnects() {
	log.Println("Test Session Ends When Client Disconnects")
	server, cancel := startServer()
	defer closeServer(server, cancel)
	time.Sleep(startupDelay)

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

	errCh := make(chan error)
	go func() {
		_, err := com.Read[com.Message](client1)
		errCh <- err
	}()
	client2.Close()
	err := <-errCh
	if err == nil {
		log.Panicf("Expected non-nil error, but received nil!")
	}
}

func assertOpponentName(start com.StartContent, expected string) {
	if start.OpponentName != expected {
		log.Panicf("Invalid opponent name. Expected: %q Was: %q", expected, start.OpponentName)
	}
}

func assertResult(result com.ResultContent, expectedOpponentSelection game.Selection, expectedResult game.Result) {
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

func startServer() (*exec.Cmd, context.CancelFunc) {
	// We want to automatically kill the server if the process jams or if it cannot be gracefully closed.
	ctx, cancel := context.WithTimeout(context.Background(), serverTimeout)

	cmd := exec.CommandContext(ctx, "./bin/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = new(syscall.SysProcAttr)
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP

	if err := cmd.Start(); err != nil {
		log.Panicf("Failed to start server process. %s", err)
	}
	return cmd, cancel
}

func closeServer(server *exec.Cmd, cancel context.CancelFunc) {
	defer cancel()
	if runtime.GOOS == "windows" {
		terminateProcess(server.Process.Pid)
		if _, err := server.Process.Wait(); err != nil {
			log.Panicf("Failed to wait process. %s", err)
		}
	} else {
		if err := server.Process.Kill(); err != nil {
			log.Panicf("Failed to kill server process. %s", err)
		}
	}
}

func newClient() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverHost, serverPort))
	if err != nil {
		log.Panicf("Failed to open TCP connection to server. %s", err)
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

// terminateProcess is a utility to send a termination signal to process in a Windows environment.
func terminateProcess(pid int) {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		log.Panicf("Failed to load kernel32.dll. %s", err)
	}
	defer func() {
		if err := dll.Release(); err != nil {
			log.Printf("Failed to release kernel32.dll. %s", err)
		}
	}()
	proc, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		log.Panicf("Failed to create console CTRL event. %s", err)
	}
	result, _, e := proc.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if result == 0 {
		log.Panicf("Failed to call break event. %s", e)
	}
}
