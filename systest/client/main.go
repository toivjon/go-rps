package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetPrefix("[system-test] ")
	log.Println("Running client system tests...")
	testReturnErrorWhenConnectingFails()
}

func testReturnErrorWhenConnectingFails() {
	log.Println("Test that client exits with error code (1) if initial connection fails.")
	client := startClient()
	state, err := client.Process.Wait()

	assertNoError(err)
	assertExited(state)
	assertExitCode(1, state.ExitCode())
}

func startClient() *exec.Cmd {
	cmd := exec.Command("./bin/client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Panicf("Failed to start client process. %s", err)
	}
	return cmd
}

func closeClient(client *exec.Cmd) {
	if err := client.Process.Kill(); err != nil {
		log.Panicf("Failed to kill client process. %s", err)
	}
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
