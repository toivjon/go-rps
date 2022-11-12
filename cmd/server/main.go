package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
)

const (
	defaultPort = 7777
	defaultHost = "localhost"
)

type Player struct {
	Conn net.Conn
	Name string
}

func main() {
	port := flag.Uint("port", defaultPort, "The port to listen for connections.")
	host := flag.String("host", defaultHost, "The network address to listen for connections.")
	flag.Parse()

	log.Println("Starting RPS server...")
	if err := start(*port, *host); err != nil {
		log.Fatal(err)
	}
}

func start(port uint, host string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()
	log.Printf("Waiting for clients on port: %d", port)

	accept := newAccept(listener)
	disconnect := make(chan net.Conn)

	conns := make(map[net.Conn]*Player)

	for {
		select {
		case conn := <-accept:
			conns[conn] = &Player{Conn: conn, Name: ""}
			log.Printf("Hello: %v (conns: %d)", conn, len(conns))
			go processConnection(conn, disconnect, conns[conn])
		case conn := <-disconnect:
			log.Printf("Bye Bye: %s", conns[conn].Name)
			delete(conns, conn)
			log.Printf("Connection %v removed (conns: %d)", conn, len(conns))
		}
	}
}

func newAccept(listener net.Listener) <-chan net.Conn {
	accept := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting incoming connection: %v", err.Error())
			} else {
				accept <- conn
			}
		}
	}()
	return accept
}

func processConnection(conn net.Conn, disconnect chan<- net.Conn, player *Player) {
	defer func() {
		conn.Close()
		disconnect <- conn
	}()

	input, err := com.Read[com.Message](conn)
	if err != nil {
		log.Printf("Failed to read data: %s", err)
		disconnect <- conn
		return
	}

	content := new(com.ConnectContent)
	if err := json.Unmarshal(input.Content, content); err != nil {
		log.Printf("Failed to unmarshal message content. %v", err)
		disconnect <- conn
	}

	player.Name = content.Name

	log.Printf("Read message: %+v content: %+v", input, content)

	if err := com.WriteConnected(conn); err != nil {
		log.Printf("Failed to write data: %s", err)
		return
	}

	log.Printf("Successfully sent response.")
}
