package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
)

type Server struct {
	conns    map[net.Conn]*Client
	joinCh   chan Message[com.JoinContent]
	selectCh chan Message[com.SelectContent]
	leaveCh  chan net.Conn
}

type Client struct {
	conn    net.Conn
	name    string
	session *Session
}

type Session struct {
	cli1          *Client
	cli2          *Client
	cli1Selection game.Selection
	cli2Selection game.Selection
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

func (s *Server) Run(port uint, host string) error {
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
	s.conns[conn] = &Client{conn: conn}
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
			content := com.JoinContent{}
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			s.joinCh <- Message[com.JoinContent]{conn: conn, content: content}
		case com.TypeSelect:
			content := com.SelectContent{}
			if err := json.Unmarshal(message.Content, &content); err != nil {
				log.Printf("Failed to unmarshal %T message content. %s", content, err)
				return
			}
			s.selectCh <- Message[com.SelectContent]{conn: conn, content: content}
		}
	}
}

func (s *Server) handleJoin(conn net.Conn, content com.JoinContent) {
	if client, ok := s.conns[conn]; ok {
		client.name = content.Name
		s.conns[conn] = client
		log.Printf("Connection %#p joined (name: %s)", conn, content.Name)

		for _, otherClient := range s.conns {
			// TODO improve the way how to detect that otherClient actually has joined! Now we use name here!
			if otherClient.session == nil && otherClient != client && otherClient.name != "" {
				if err := s.startSession(client, otherClient); err != nil {
					log.Printf("Failed to start session for connection %#p and %#p", client.conn, otherClient.conn)
				}
			}
		}
	}
}

func (s *Server) startSession(cli1, cli2 *Client) error {
	session := &Session{cli1: cli1, cli2: cli2}
	if err := com.WriteMessage(cli1.conn, com.TypeStart, com.StartContent{OpponentName: cli2.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", cli1.conn, err)
	}
	if err := com.WriteMessage(cli2.conn, com.TypeStart, com.StartContent{OpponentName: cli1.name}); err != nil {
		return fmt.Errorf("failed to write START message for conn %#p. %w", cli2.conn, err)
	}
	cli1.session = session
	cli2.session = session
	s.conns[cli1.conn] = cli1
	s.conns[cli2.conn] = cli2
	log.Printf("Session %#p started (conn1: %#p conn2: %#p)", session, cli1.conn, cli2.conn)
	return nil
}

func (s *Server) handleSelect(conn net.Conn, content com.SelectContent) error {
	if client, ok := s.conns[conn]; ok {
		log.Printf("Connection %#p selection received (selection: %s)", conn, content.Selection)
		session := client.session
		switch client.conn {
		case session.cli1.conn:
			session.cli1Selection = content.Selection
		case session.cli2.conn:
			session.cli2Selection = content.Selection
		}
		selection1 := session.cli1Selection
		selection2 := session.cli2Selection
		if selection1 != game.SelectionNone && selection2 != game.SelectionNone {
			result1 := game.ResultDraw
			result2 := game.ResultDraw
			switch {
			case selection1 == selection2:
				break
			case selection1.Beats(selection2):
				result1 = game.ResultWin
				result2 = game.ResultLose
			default:
				result1 = game.ResultLose
				result2 = game.ResultWin
			}
			conn1 := session.cli1.conn
			conn2 := session.cli2.conn
			messageContent := com.ResultContent{OpponentSelection: selection2, Result: result1}
			if err := com.WriteMessage(conn1, com.TypeResult, messageContent); err != nil {
				return fmt.Errorf("failed to write RESULT message for conn %#p. %w", conn1, err)
			}
			messageContent = com.ResultContent{OpponentSelection: selection1, Result: result2}
			if err := com.WriteMessage(conn2, com.TypeResult, messageContent); err != nil {
				return fmt.Errorf("failed to write RESULT message for conn  %#p. %w", conn2, err)
			}
			log.Printf("Session %#p round result %#p:%s and %#p:%s", session, conn1, result1, conn2, result2)
			if result1 == game.ResultDraw && result2 == game.ResultDraw {
				session.cli1Selection = game.SelectionNone
				session.cli2Selection = game.SelectionNone
			}
		}
	}
	return nil
}

func (s *Server) handleLeave(conn net.Conn) {
	if client, ok := s.conns[conn]; ok {
		delete(s.conns, conn)
		if client.session != nil {
			s.closeSession(client.session)
		}
		log.Printf("Connection %#p removed (conns: %d).", conn, len(s.conns))
	}
}

func (s *Server) closeSession(session *Session) {
	session.cli1.session = nil
	session.cli2.session = nil
	log.Printf("Session %#p closed (conn1: %#p conn2: %#p)", session, &session.cli1.conn, &session.cli2.conn)
	session.cli1.conn.Close()
	session.cli2.conn.Close()
}
