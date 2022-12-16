package server

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/toivjon/go-rps/internal/com"
)

// Server represents a RPS server handling the connection communication, matchmaking and game logics.
type Server struct {
	listener net.Listener
	conns    map[io.ReadWriteCloser]*Client
	joinCh   chan Message[com.JoinContent]
	selectCh chan Message[com.SelectContent]
	leaveCh  chan io.ReadWriteCloser
	shutdown <-chan os.Signal
}

type Message[T any] struct {
	conn    io.ReadWriteCloser
	content T
}

// NewServer builds a new server with the given network listener and shutdown channel.
func NewServer(listener net.Listener, shutdown <-chan os.Signal) Server {
	return Server{
		listener: listener,
		conns:    make(map[io.ReadWriteCloser]*Client),
		joinCh:   make(chan Message[com.JoinContent]),
		selectCh: make(chan Message[com.SelectContent]),
		leaveCh:  make(chan io.ReadWriteCloser),
		shutdown: shutdown,
	}
}

func (s *Server) Run() error {
	accept := newAccept(s.listener)
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
		case <-s.shutdown:
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
