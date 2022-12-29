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
	Listener net.Listener
	Conns    map[io.ReadWriteCloser]*Client
	JoinCh   chan Message[com.JoinContent]
	SelectCh chan Message[com.SelectContent]
	LeaveCh  chan io.ReadWriteCloser
	Shutdown <-chan os.Signal
}

// Message represents an incoming message from a client connection.
type Message[T any] struct {
	Conn    io.ReadWriteCloser
	Content T
}

// NewServer builds a new server with the given network listener and shutdown channel.
func NewServer(listener net.Listener, shutdown <-chan os.Signal) Server {
	return Server{
		Listener: listener,
		Conns:    make(map[io.ReadWriteCloser]*Client),
		JoinCh:   make(chan Message[com.JoinContent]),
		SelectCh: make(chan Message[com.SelectContent]),
		LeaveCh:  make(chan io.ReadWriteCloser),
		Shutdown: shutdown,
	}
}

// Run starts running the server main loop which accepts new connections and handles incoming messages.
func (s *Server) Run() {
	accept := newAccept(s.Listener)
	for {
		select {
		case conn := <-accept:
			s.handleAccept(conn)
		case message := <-s.JoinCh:
			s.handleJoin(message.Conn, message.Content)
		case message := <-s.SelectCh:
			s.handleSelect(message.Conn, message.Content)
		case conn := <-s.LeaveCh:
			s.handleLeave(conn)
		case <-s.Shutdown:
			log.Printf("Shutting down server...")
			// ...
			return
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
	s.Conns[conn] = client
	go client.Run(s.LeaveCh, s.JoinCh, s.SelectCh)
	log.Printf("Connection %#p added (conns: %d).", conn, len(s.Conns))
}

func (s *Server) handleJoin(conn io.ReadWriteCloser, content com.JoinContent) {
	if client, ok := s.Conns[conn]; ok {
		client.Name = content.Name
		s.Conns[conn] = client
		log.Printf("Connection %#p joined (name: %s)", conn, content.Name)

		for _, otherClient := range s.Conns {
			// ... improve the way how to detect that otherClient actually has joined! Now we use name here!
			if otherClient.Session == nil && otherClient != client && otherClient.Name != "" {
				session := NewSession(client, otherClient)
				if err := session.Start(); err != nil {
					log.Printf("Failed to start session for connection %s and %s", client, otherClient)
					otherClient.Session = nil
					client.Session = nil
				}
			}
		}
	}
}

func (s *Server) handleSelect(conn io.ReadWriteCloser, content com.SelectContent) {
	if client, ok := s.Conns[conn]; ok {
		log.Printf("Connection %#p selection received (selection: %s)", conn, content.Selection)
		if err := client.Session.Select(client, content.Selection); err != nil {
			log.Printf("Failed to process SELECT in session %#p. %s", client.Session, err)
			client.Session.Close()
		}
	}
}

func (s *Server) handleLeave(conn io.ReadWriteCloser) {
	if client, ok := s.Conns[conn]; ok {
		delete(s.Conns, conn)
		if client.Session != nil {
			client.Session.Close()
		}
		log.Printf("Connection %#p removed (conns: %d).", conn, len(s.Conns))
	}
}
