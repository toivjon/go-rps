package client

import (
	"fmt"
	"log"
	"net"
)

func Start(port uint, host, name string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	cli := NewClient(conn, name)
	if err := cli.Run(); err != nil {
		return fmt.Errorf("failed to run client. %w", err)
	}
	return nil
}
