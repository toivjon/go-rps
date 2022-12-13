package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
)

type Server struct {
	conns    map[net.Conn]*Client
	joinCh   chan Message[com.JoinContent]
	selectCh chan Message[com.SelectContent]
	leaveCh  chan net.Conn
}

type Message[T any] struct {
	conn    net.Conn
	content T
}

func NewServer() Server {
	return Server{
		conns:    make(map[net.Conn]*Client),
		joinCh:   make(chan Message[com.JoinContent]),
		selectCh: make(chan Message[com.SelectContent]),
		leaveCh:  make(chan net.Conn),
	}
}

func (s *Server) Run(port uint, host string, shutdown chan os.Signal) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start listening TCP socket on port %d. %w", port, err)
	}
	defer listener.Close()

	log.Printf("Waiting for clients on port: %d", port)
	accept := newAccept(listener)
	for {
		select {
		case conn := <-accept:
			s.handleAccept(conn)
		case message := <-s.joinCh:
			s.handleJoin(message.conn, message.content)
		case message := <-s.selectCh:
			s.handleSelect(message.conn, message.content)
		case conn := <-s.leaveCh:
			s.handleLeave(conn)
		case <-shutdown:
			log.Printf("Shutting down server...")
			// ...
			return nil
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

func (s *Server) handleAccept(conn net.Conn) {
	s.conns[conn] = NewClient(conn)
	go s.processClient(conn)
	log.Printf("Connection %#p added (conns: %d).", conn, len(s.conns))
}

func (s *Server) processClient(conn net.Conn) {
	defer func() {
		s.leaveCh <- conn
		conn.Close()
	}()
	for {
		message, err := com.Read[com.Message](conn)
		if err != nil {
			return
		}
		switch message.Type {
		case com.TypeJoin:
			content := new(com.JoinContent)
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			s.joinCh <- Message[com.JoinContent]{conn: conn, content: *content}
		case com.TypeSelect:
			content := new(com.SelectContent)
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			s.selectCh <- Message[com.SelectContent]{conn: conn, content: *content}
		case com.TypeResult, com.TypeStart:
			log.Printf("Connection %#p received unsupported message type %s!", conn, message.Type)
			return
		}
	}
}

func (s *Server) handleJoin(conn net.Conn, content com.JoinContent) {
	if client, ok := s.conns[conn]; ok {
		client.name = content.Name
		s.conns[conn] = client
		log.Printf("Connection %#p joined (name: %s)", conn, content.Name)

		for _, otherClient := range s.conns {
			// ... improve the way how to detect that otherClient actually has joined! Now we use name here!
			if otherClient.session == nil && otherClient != client && otherClient.name != "" {
				session := NewSession(client, otherClient)
				if err := session.Start(); err != nil {
					log.Printf("Failed to start session for connection %#p and %#p", client.conn, otherClient.conn)
				}
			}
		}
	}
}

func (s *Server) handleSelect(conn net.Conn, content com.SelectContent) {
	if client, ok := s.conns[conn]; ok {
		log.Printf("Connection %#p selection received (selection: %s)", conn, content.Selection)
		if err := client.session.Select(client, content.Selection); err != nil {
			log.Printf("Failed to process SELECT in session %#p. %s", client.session, err)
			client.session.Close()
		}
	}
}

func (s *Server) handleLeave(conn net.Conn) {
	if client, ok := s.conns[conn]; ok {
		delete(s.conns, conn)
		if client.session != nil {
			client.session.Close()
		}
		log.Printf("Connection %#p removed (conns: %d).", conn, len(s.conns))
	}
}
