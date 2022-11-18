package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
	defaultName = "anonymous"
)

func main() {
	port := flag.Uint("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	name := flag.String("name", defaultName, "The name of the player.")
	flag.Parse()

	log.Println("Starting RPS client...")
	if err := start(*port, *host, *name); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}

func start(port uint, host, name string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	if err := sendJoin(conn, name); err != nil {
		return err
	}

	log.Printf("waiting for START message...")
	startContent, err := readStart(conn)
	if err != nil {
		return err
	}
	log.Printf("starting the game: %+v", startContent)

	// ... get input from keyboard input.
	// ... wait for result.

	return nil
}

func sendJoin(writer io.Writer, name string) error {
	content, err := json.Marshal(com.JoinContent{Name: name})
	if err != nil {
		return fmt.Errorf("failed marshal JOIN content into JSON. %w", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeJoin, Content: content}); err != nil {
		return fmt.Errorf("failed to write JOIN message to connection. %w", err)
	}
	return nil
}

func readStart(reader io.Reader) (com.StartContent, error) {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		return com.StartContent{}, fmt.Errorf("failed to read START message. %w", err)
	}
	content := com.StartContent{OpponentName: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.StartContent{}, fmt.Errorf("failed to read START content. %w", err)
	}
	return content, nil
}
