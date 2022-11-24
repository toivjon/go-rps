package main

import (
	"bufio"
	"encoding/json"
	"errors"
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

var errUnexpectedMessageType = errors.New("unexpected message type")

func main() {
	port := flag.Uint("port", defaultPort, "The port of the server.")
	host := flag.String("host", defaultHost, "The IP address or hostname of the server.")
	name := flag.String("name", defaultName, "The name of the player.")
	flag.Parse()

	log.Println("Welcome to the RPS client")
	if err := start(*port, *host, *name); err != nil {
		log.Fatalf("Client was closed due an error: %v", err)
	}
	log.Println("Client was closed successfully.")
}

func start(port uint, host, name string) error {
	log.Printf("Connecting to the server: %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection: %w", err)
	}
	defer conn.Close()

	log.Printf("Joining as player: %s", name)
	if err := send(conn, com.TypeJoin, com.JoinContent{Name: name}); err != nil {
		return err
	}

	log.Printf("Waiting for an opponent. Please stand by...")
	startContent, err := read[com.StartContent](conn, com.TypeStart)
	if err != nil {
		return err
	}
	log.Printf("Opponent joined the game: %s", startContent.OpponentName)

	gameOver := false
	for !gameOver {
		if err := processSelect(conn); err != nil {
			return err
		}

		log.Println("Waiting for game result. Please wait...")
		resultContent, err := read[com.ResultContent](conn, com.TypeResult)
		if err != nil {
			return err
		}
		log.Printf("Result: %+v", resultContent)

		if resultContent.Result != game.ResultDraw {
			gameOver = true
		} else {
			log.Printf("Round ended in a draw. Let's have an another round...")
		}
	}
	log.Printf("Game over.")
	return nil
}

func processSelect(conn net.Conn) error {
	log.Println("Please type the selection ('r', 'p', 's') and press enter:")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to scan user input. %w", scanner.Err())
	}
	input := scanner.Text()

	selection := game.Selection(input)
	if err := game.ValidateSelection(selection); err != nil {
		return fmt.Errorf("failed to validate user input. %w", err)
	}

	if err := send(conn, com.TypeSelect, com.SelectContent{Selection: selection}); err != nil {
		return err
	}
	return nil
}

func send[T any](writer io.Writer, messageType com.MessageType, val T) error {
	content, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed marshal %s content into JSON. %w", messageType, err)
	}
	if err := com.Write(writer, com.Message{Type: messageType, Content: content}); err != nil {
		return fmt.Errorf("failed to write %s message to connection. %w", messageType, err)
	}
	return nil
}

func read[T any](reader io.Reader, messageType com.MessageType) (*T, error) {
	message, err := com.Read[com.Message](reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s message. %w", messageType, err)
	}
	if messageType != message.Type {
		return nil, fmt.Errorf("failed to read %s message as %s. %w", message.Type, messageType, errUnexpectedMessageType)
	}
	content := new(T)
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to read %s content. %w", messageType, err)
	}
	return content, nil
}
