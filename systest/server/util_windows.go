//go:build windows

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
)

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
	result, _, e := proc.Call(syscall.CTRL_BREAK_EVENT, uintptr(server.Process.Pid))
	if result == 0 {
		log.Panicf("Failed to call break event. %s", e)
	}
	if _, err := server.Process.Wait(); err != nil {
		log.Panicf("Failed to wait process. %s", err)
	}
}
