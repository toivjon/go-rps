package server_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/toivjon/go-rps/internal/com"
	"github.com/toivjon/go-rps/internal/game"
	"github.com/toivjon/go-rps/internal/server"
)

var errMock = errors.New("error-mock")

type readerResult struct {
	data []byte
	err  error
}

type readerMock struct {
	results []readerResult
}

func (r *readerMock) Read(p []byte) (int, error) {
	result := r.results[0]
	copy(p, result.data)
	r.results = r.results[1:]
	return len(result.data), result.err
}

type writerMock struct {
	n   int
	err error
}

func (w *writerMock) Write(p []byte) (int, error) {
	return w.n, w.err
}

type closerMock struct {
	err error
}

func (c *closerMock) Close() error {
	return c.err
}

type connMock struct {
	readerMock
	writerMock
	closerMock
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	conn := new(connMock)
	cli := server.NewClient(conn)
	if cli.Conn != conn {
		t.Fatalf("Expected to contain conn member %#p but had %#p", conn, cli.Conn)
	}
	if len(cli.Name) != 0 {
		t.Fatalf("Expected to contain empty name member but had %s", cli.Name)
	}
	if cli.Session != nil {
		t.Fatalf("Expected to contain nil session but had %#p", cli.Session)
	}
}

func TestClientWriteStart(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenWriteFails", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		conn.writerMock.err = errMock
		cli := server.NewClient(conn)
		if err := cli.WriteStart(""); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnNilWhenSuccess", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		cli := server.NewClient(conn)
		if err := cli.WriteStart(""); err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}

func TestClientWriteResult(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenWriteFails", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		conn.writerMock.err = errMock
		cli := server.NewClient(conn)
		if err := cli.WriteResult(game.SelectionRock, game.ResultWin); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnNilWhenSuccess", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		cli := server.NewClient(conn)
		if err := cli.WriteResult(game.SelectionRock, game.ResultWin); err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}

//nolint:funlen,cyclop
func TestClientRun(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenReadFails", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: nil, err: errMock})
		cli := server.NewClient(conn)
		leaveCh := make(chan io.ReadWriteCloser, 1)
		cli.Run(leaveCh, nil, nil)
		if leaveConn := <-leaveCh; leaveConn != conn {
			t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
		}
	})
	t.Run("ReturnErrorWhenJoinUnmarshalFails", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"JOIN","content":"non-json"}`
		conn := new(connMock)
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: []byte(data), err: nil})
		cli := server.NewClient(conn)
		leaveCh := make(chan io.ReadWriteCloser, 1)
		cli.Run(leaveCh, nil, nil)
		if leaveConn := <-leaveCh; leaveConn != conn {
			t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
		}
	})
	t.Run("ReturnErrorWhenSelectUnmarshalFails", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"SELECT","content":"non-json"}`
		conn := new(connMock)
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: []byte(data), err: nil})
		cli := server.NewClient(conn)
		leaveCh := make(chan io.ReadWriteCloser, 1)
		cli.Run(leaveCh, nil, nil)
		if leaveConn := <-leaveCh; leaveConn != conn {
			t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
		}
	})
	t.Run("ReturnErrorWhenUnsupportedTypeIsReceived", func(t *testing.T) {
		t.Parallel()
		for _, messageType := range []com.MessageType{com.TypeResult, com.TypeStart} {
			data := fmt.Sprintf(`{"type":"%s","content":{}}`, messageType)
			conn := new(connMock)
			conn.readerMock.results = append(conn.readerMock.results, readerResult{data: []byte(data), err: nil})
			cli := server.NewClient(conn)
			leaveCh := make(chan io.ReadWriteCloser, 1)
			cli.Run(leaveCh, nil, nil)
			if leaveConn := <-leaveCh; leaveConn != conn {
				t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
			}
		}
	})
	t.Run("CallJoinChannelWhenJoinMessageIsReceived", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"JOIN","content":{"name":"donald"}}`
		conn := new(connMock)
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: []byte(data), err: nil})
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: nil, err: errMock})
		cli := server.NewClient(conn)
		leaveCh := make(chan io.ReadWriteCloser, 1)
		joinCh := make(chan server.Message[com.JoinContent], 1)
		cli.Run(leaveCh, joinCh, nil)
		joinCall := <-joinCh
		if joinCall.Conn != conn {
			t.Fatalf("Expected join call to contain connection %#p but had %#p!", conn, joinCall.Conn)
		}
		if joinCall.Content.Name != "donald" {
			t.Fatalf("Expected join call to contain name \"donald\" but had %q!", joinCall.Content.Name)
		}
		if leaveConn := <-leaveCh; leaveConn != conn {
			t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
		}
	})
	t.Run("CallSelectChannelWhenSelectMessageIsReceived", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"SELECT","content":{"selection":"r"}}`
		conn := new(connMock)
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: []byte(data), err: nil})
		conn.readerMock.results = append(conn.readerMock.results, readerResult{data: nil, err: errMock})
		cli := server.NewClient(conn)
		leaveCh := make(chan io.ReadWriteCloser, 1)
		selectCh := make(chan server.Message[com.SelectContent], 1)
		cli.Run(leaveCh, nil, selectCh)
		selectCall := <-selectCh
		if selectCall.Conn != conn {
			t.Fatalf("Expected join call to contain connection %#p but had %#p!", conn, selectCall.Conn)
		}
		if selectCall.Content.Selection != game.SelectionRock {
			t.Fatalf("Expected join call to contain name r but had %q!", selectCall.Content.Selection)
		}
		if leaveConn := <-leaveCh; leaveConn != conn {
			t.Fatalf("Expected leave to be called with connection %#p but was %#p!", conn, leaveConn)
		}
	})
}

func TestString(t *testing.T) {
	t.Parallel()
	conn := new(connMock)
	cli := server.Client{Conn: conn, Name: "foo", Session: nil}
	expected := fmt.Sprintf("client(%#p:%s)", conn, "foo")
	if val := cli.String(); val != expected {
		t.Fatalf("Expected to return %s but returned %q!", expected, val)
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenCloseFails", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		conn.closerMock.err = errMock
		cli := server.NewClient(conn)
		if err := cli.Close(); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnNilWhenCloseSucceeds", func(t *testing.T) {
		t.Parallel()
		conn := new(connMock)
		cli := server.NewClient(conn)
		if err := cli.Close(); err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}
