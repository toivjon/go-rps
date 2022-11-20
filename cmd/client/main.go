package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
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

	gameOver := false
	for !gameOver {
		log.Println("Enter selection ('r', 'p', 's')")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("failed to scan user input. %w", err)
		}
		input := scanner.Text()

		selection := game.Selection(input)
		if err := game.ValidateSelection(selection); err != nil {
			return fmt.Errorf("failed to validate user input. %w", err)
		}

		if err := sendSelect(conn, selection); err != nil {
			return err
		}

		log.Println("Waiting for game result. Please wait...")
		resultContent, err := readResult(conn)
		if err != nil {
			return err
		}
		log.Printf("Result: %+v", resultContent)

		if resultContent.Result != "DRAW" {
			gameOver = true
		} else {
			log.Printf("Round ended in a draw. Let's have an another round...")
		}
	}
	log.Printf("Game over.")
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

func sendSelect(writer io.Writer, selection game.Selection) error {
	content, err := json.Marshal(com.SelectContent{Selection: selection})
	if err != nil {
		return fmt.Errorf("failed to marshal SELECT content into JSON. %w", err)
	}
	if err := com.Write(writer, com.Message{Type: com.TypeSelect, Content: content}); err != nil {
		return fmt.Errorf("failed to write SELECT message to connection. %w", err)
	}
	return nil
}

func readResult(reader io.Reader) (com.ResultContent, error) {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		return com.ResultContent{}, fmt.Errorf("failed to read RESULT message. %w", err)
	}
	content := com.ResultContent{OpponentSelection: "", Result: ""}
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return com.ResultContent{}, fmt.Errorf("failed to read RESULT content. %w", err)
	}
	return content, nil
}
