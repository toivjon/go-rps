package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
)

type Server struct {
	conns    map[io.ReadWriteCloser]*Client
	joinCh   chan Message[com.JoinContent]
	selectCh chan Message[com.SelectContent]
	leaveCh  chan io.ReadWriteCloser
}

type Message[T any] struct {
	conn    io.ReadWriteCloser
	content T
}

func NewServer() Server {
	return Server{
		conns:    make(map[io.ReadWriteCloser]*Client),
		joinCh:   make(chan Message[com.JoinContent]),
		selectCh: make(chan Message[com.SelectContent]),
		leaveCh:  make(chan io.ReadWriteCloser),
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
				log.Printf("Error accepting incoming connection: %s", err)
			} else {
				accept <- conn
			}
		}
	}()
	return accept
}

func (s *Server) handleAccept(conn io.ReadWriteCloser) {
	client := NewClient(conn)
	s.conns[conn] = client
	go client.Run(s)
	log.Printf("Connection %#p added (conns: %d).", conn, len(s.conns))
}

func (s *Server) handleJoin(conn io.ReadWriteCloser, content com.JoinContent) {
	if client, ok := s.conns[conn]; ok {
		client.name = content.Name
		s.conns[conn] = client
		log.Printf("Connection %#p joined (name: %s)", conn, content.Name)

		for _, otherClient := range s.conns {
			// ... improve the way how to detect that otherClient actually has joined! Now we use name here!
			if otherClient.session == nil && otherClient != client && otherClient.name != "" {
				session := NewSession(client, otherClient)
				if err := session.Start(); err != nil {
					log.Printf("Failed to start session for connection %s and %s", client, otherClient)
				}
			}
		}
	}
}

func (s *Server) handleSelect(conn io.ReadWriteCloser, content com.SelectContent) {
	if client, ok := s.conns[conn]; ok {
		log.Printf("Connection %#p selection received (selection: %s)", conn, content.Selection)
		if err := client.session.Select(client, content.Selection); err != nil {
			log.Printf("Failed to process SELECT in session %#p. %s", client.session, err)
			client.session.Close()
		}
	}
}

func (s *Server) handleLeave(conn io.ReadWriteCloser) {
	if client, ok := s.conns[conn]; ok {
		delete(s.conns, conn)
		if client.session != nil {
			client.session.Close()
		}
		log.Printf("Connection %#p removed (conns: %d).", conn, len(s.conns))
	}
}
