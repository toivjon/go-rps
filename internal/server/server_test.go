package server_test

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
	"github.com/toivjon/go-rps/internal/server"
)

type listenerMock struct {
	acceptErr error
	acceptCh  chan net.Conn
}

func (l *listenerMock) Accept() (net.Conn, error) {
	if l.acceptErr != nil {
		err := l.acceptErr
		l.acceptErr = nil
		return nil, err
	}
	return <-l.acceptCh, nil
}

func (l *listenerMock) Close() error {
	return nil
}

func (l *listenerMock) Addr() net.Addr {
	return new(net.IPAddr)
}

type fullConnMock struct {
	readCh   chan any
	writeErr error
}

func (f *fullConnMock) Read(b []byte) (int, error) {
	<-f.readCh
	return 0, nil
}

func (f *fullConnMock) Write(b []byte) (int, error) {
	return 0, f.writeErr
}

func (f *fullConnMock) Close() error {
	return nil
}

func (f *fullConnMock) LocalAddr() net.Addr {
	return new(net.IPAddr)
}

func (f *fullConnMock) RemoteAddr() net.Addr {
	return new(net.IPAddr)
}

func (f *fullConnMock) SetDeadline(t time.Time) error {
	return nil
}

func (f *fullConnMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *fullConnMock) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestNewServer(t *testing.T) {
	t.Parallel()
	listenerMock := new(listenerMock)
	shutdown := make(chan os.Signal)
	server := server.NewServer(listenerMock, shutdown)
	if server.Listener != listenerMock {
		t.Fatalf("Expected listener member to be %#p but was %#p!", listenerMock, server.Listener)
	}
	if server.Shutdown != shutdown {
		t.Fatalf("Expected shutdown channel to be %#p but it was %#p!", &shutdown, server.Shutdown)
	}
}

//nolint:funlen,cyclop
func TestServerRun(t *testing.T) {
	t.Parallel()
	t.Run("SkipFailedAccept", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptErr = errMock
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		server := server.NewServer(listenerMock, shutdown)
		go server.Run()
		shutdown <- os.Kill
		if len(server.Conns) != 0 {
			t.Fatalf("Expected connections to be empty, but was %v!", server.Conns)
		}
	})
	t.Run("StartNewClientOnAccept", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		server := server.NewServer(listenerMock, shutdown)
		go server.Run()
		conn := new(fullConnMock)
		listenerMock.acceptCh <- conn
		time.Sleep(time.Second)
		if len(server.Conns) != 1 {
			t.Fatalf("Expected connections to contain one item, but had %d!", len(server.Conns))
		}
		if server.Conns[conn].Conn != conn {
			t.Fatal("Expected client to wrap connection, but it did not!")
		}
		shutdown <- os.Kill
	})
	t.Run("UpdateClientNameOnJoin", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()
		conn := new(fullConnMock)
		srv.Conns[conn] = server.NewClient(conn)
		srv.JoinCh <- server.Message[com.JoinContent]{Conn: conn, Content: com.JoinContent{Name: "donald"}}
		time.Sleep(time.Second)
		if srv.Conns[conn].Name != "donald" {
			t.Fatalf("Expected client to have name \"donald\", but had %q!", srv.Conns[conn].Name)
		}
		shutdown <- os.Kill
	})
	t.Run("StartSessionOnMatchmakeDuringJoin", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn1 := new(fullConnMock)
		srv.Conns[conn1] = server.NewClient(conn1)
		srv.JoinCh <- server.Message[com.JoinContent]{Conn: conn1, Content: com.JoinContent{Name: "donald"}}
		time.Sleep(time.Second)

		conn2 := new(fullConnMock)
		srv.Conns[conn2] = server.NewClient(conn2)
		srv.JoinCh <- server.Message[com.JoinContent]{Conn: conn2, Content: com.JoinContent{Name: "mickey"}}
		time.Sleep(time.Second)

		session := srv.Conns[conn1].Session
		if session == nil {
			t.Fatal("Expected client session to be non-nil, but was nil!")
		}
		if session != srv.Conns[conn2].Session {
			t.Fatal("Expected clients to contain same session, but did not!")
		}
		shutdown <- os.Kill
	})
	t.Run("SkipFailedSessionStart", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn1 := new(fullConnMock)
		srv.Conns[conn1] = server.NewClient(conn1)
		srv.JoinCh <- server.Message[com.JoinContent]{Conn: conn1, Content: com.JoinContent{Name: "donald"}}
		time.Sleep(time.Second)

		conn2 := new(fullConnMock)
		conn2.writeErr = errMock
		srv.Conns[conn2] = server.NewClient(conn2)
		srv.JoinCh <- server.Message[com.JoinContent]{Conn: conn2, Content: com.JoinContent{Name: "mickey"}}
		time.Sleep(time.Second)

		if srv.Conns[conn1].Session != nil {
			t.Fatal("Expected client1 session to nil!")
		}
		if srv.Conns[conn2].Session != nil {
			t.Fatal("Expected client2 session to nil!")
		}
		shutdown <- os.Kill
	})
	t.Run("PerformSelectOnSelect", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn := new(fullConnMock)
		srv.Conns[conn] = server.NewClient(conn)
		srv.Conns[conn].Session = &server.Session{
			Cli1:  srv.Conns[conn],
			Cli2:  srv.Conns[conn],
			Round: server.NewRound(),
		}
		srv.SelectCh <- server.Message[com.SelectContent]{
			Conn:    conn,
			Content: com.SelectContent{Selection: game.SelectionRock},
		}
		time.Sleep(time.Second)

		round := srv.Conns[conn].Session.Round
		if round.Selection1 != game.SelectionRock {
			t.Fatal("Expcted selection1 to be rock!")
		}
		if round.Selection2 != game.SelectionNone {
			t.Fatal("Expcted selection1 to be none!")
		}
		shutdown <- os.Kill
	})
	t.Run("CloseSessionOnFailedSelect", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn1 := new(fullConnMock)
		conn1.writeErr = errMock
		conn2 := new(fullConnMock)
		srv.Conns[conn1] = server.NewClient(conn1)
		srv.Conns[conn2] = server.NewClient(conn2)
		srv.Conns[conn1].Session = &server.Session{
			Cli1: srv.Conns[conn1],
			Cli2: srv.Conns[conn2],
			Round: &server.Round{
				Selection1: game.SelectionNone,
				Selection2: game.SelectionRock,
			},
		}
		srv.SelectCh <- server.Message[com.SelectContent]{
			Conn:    conn1,
			Content: com.SelectContent{Selection: game.SelectionRock},
		}
		time.Sleep(time.Second)

		if srv.Conns[conn1].Session != nil {
			t.Fatal("Expected conn1 session to be closed and nil!")
		}
		if srv.Conns[conn2].Session != nil {
			t.Fatal("Expected conn2 session to be closed and nil!")
		}
		shutdown <- os.Kill
	})
	t.Run("RemoveConnectionOnLeave", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn := new(fullConnMock)
		srv.Conns[conn] = server.NewClient(conn)
		srv.LeaveCh <- conn
		time.Sleep(time.Second)

		if len(srv.Conns) != 0 {
			t.Fatalf("Expected connections list to be empty, but was %v!", srv.Conns)
		}
		shutdown <- os.Kill
	})
	t.Run("CloseSessionOnLeave", func(t *testing.T) {
		t.Parallel()
		listenerMock := new(listenerMock)
		listenerMock.acceptCh = make(chan net.Conn)
		shutdown := make(chan os.Signal)
		srv := server.NewServer(listenerMock, shutdown)
		go srv.Run()

		conn1 := new(fullConnMock)
		conn2 := new(fullConnMock)
		srv.Conns[conn1] = server.NewClient(conn1)
		srv.Conns[conn2] = server.NewClient(conn2)
		server.NewSession(srv.Conns[conn1], srv.Conns[conn2])
		srv.LeaveCh <- conn1
		time.Sleep(time.Second)

		if len(srv.Conns) != 1 {
			t.Fatalf("Expected connections list to contain one item, but had %v!", srv.Conns)
		}
		if srv.Conns[conn2].Session != nil {
			t.Fatalf("Expected conn2 to contain nil session, but had %v!", srv.Conns[conn2].Session)
		}
		shutdown <- os.Kill
	})
}
