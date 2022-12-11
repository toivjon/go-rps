//go:build linux

package main

func startServer() (*exec.Cmd, context.CancelFunc) {
	// We want to automatically kill the server if the process jams or if it cannot be gracefully closed.
	ctx, cancel := context.WithTimeout(context.Background(), serverTimeout)

	cmd := exec.CommandContext(ctx, "./bin/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Panicf("Failed to start server process. %s", err)
	}
	return cmd, cancel
}

func closeServer(server *exec.Cmd, cancel context.CancelFunc) {
	defer cancel()
	if err := server.Process.Kill(); err != nil {
		log.Panicf("Failed to kill server process. %s", err)
	}
}
