package server_test

import (
	"errors"
	"testing"

	"github.com/toivjon/go-rps/internal/game"
	"github.com/toivjon/go-rps/internal/server"
)

func TestNewSession(t *testing.T) {
	t.Parallel()
	cli1 := server.NewClient(new(connMock))
	cli2 := server.NewClient(new(connMock))
	session := server.NewSession(cli1, cli2)
	if session.Cli1 != cli1 {
		t.Fatalf("Expected cli1 to be %#p, but was %#p!", cli1, session.Cli1)
	}
	if session.Cli2 != cli2 {
		t.Fatalf("Expected cli2 to be %#p, but was %#p!", cli2, session.Cli2)
	}
	if session.Round == nil {
		t.Fatal("Expected round to be non-nil, but was nil!")
	}
}

func TestSessionStart(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenCli1WriteFails", func(t *testing.T) {
		t.Parallel()
		errConn := new(connMock)
		errConn.writerMock.err = errMock
		cli1 := server.NewClient(errConn)
		cli2 := server.NewClient(new(connMock))
		session := server.NewSession(cli1, cli2)
		if err := session.Start(); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnErrorWhenCli2WriteFails", func(t *testing.T) {
		t.Parallel()
		errConn := new(connMock)
		errConn.writerMock.err = errMock
		cli1 := server.NewClient(new(connMock))
		cli2 := server.NewClient(errConn)
		session := server.NewSession(cli1, cli2)
		if err := session.Start(); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnNilWhenWritesSucceed", func(t *testing.T) {
		t.Parallel()
		cli1 := server.NewClient(new(connMock))
		cli2 := server.NewClient(new(connMock))
		session := server.NewSession(cli1, cli2)
		if err := session.Start(); err != nil {
			t.Fatalf("Expected no error, but an error %q was returned!", err)
		}
	})
}

//nolint:funlen,cyclop
func TestSessionSelect(t *testing.T) {
	t.Parallel()
	t.Run("NoFurtherActionsWhenSelection1StaysNone", func(t *testing.T) {
		t.Parallel()
		conn1 := new(connMock)
		conn2 := new(connMock)
		conn1.writerMock.err = errMock
		conn2.writerMock.err = errMock
		cli1 := server.NewClient(conn1)
		cli2 := server.NewClient(conn2)
		session := server.NewSession(cli1, cli2)
		if err := session.Select(cli1, game.SelectionRock); err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
		if session.Round.Selection1 != game.SelectionRock {
			t.Fatal("Expected selection1 to be assigned, but it was not!")
		}
		if session.Round.Selection2 != game.SelectionNone {
			t.Fatalf("Expected selection2 to still be none, but it was %q!", session.Round.Selection2)
		}
	})
	t.Run("NoFurtherActionsWhenSelection2StaysNone", func(t *testing.T) {
		t.Parallel()
		conn1 := new(connMock)
		conn2 := new(connMock)
		conn1.writerMock.err = errMock
		conn2.writerMock.err = errMock
		cli1 := server.NewClient(conn1)
		cli2 := server.NewClient(conn2)
		session := server.NewSession(cli1, cli2)
		if err := session.Select(cli2, game.SelectionRock); err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
		if session.Round.Selection1 != game.SelectionNone {
			t.Fatalf("Expected selection1 to still be none, but it was %q!", session.Round.Selection1)
		}
		if session.Round.Selection2 != game.SelectionRock {
			t.Fatal("Expected selection1 to be assigned, but it was not!")
		}
	})
	t.Run("ReturnErrorWhenCli1WriteFails", func(t *testing.T) {
		t.Parallel()
		errConn := new(connMock)
		errConn.writerMock.err = errMock
		cli1 := server.NewClient(errConn)
		cli2 := server.NewClient(new(connMock))
		session := server.NewSession(cli1, cli2)
		session.Round.Selection2 = game.SelectionRock
		if err := session.Select(cli1, game.SelectionRock); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnErrorWhenCli2WriteFails", func(t *testing.T) {
		t.Parallel()
		errConn := new(connMock)
		errConn.writerMock.err = errMock
		cli1 := server.NewClient(new(connMock))
		cli2 := server.NewClient(errConn)
		session := server.NewSession(cli1, cli2)
		session.Round.Selection2 = game.SelectionRock
		if err := session.Select(cli1, game.SelectionRock); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("StartNewRoundOnDraw", func(t *testing.T) {
		t.Parallel()
		cli1 := server.NewClient(new(connMock))
		cli2 := server.NewClient(new(connMock))
		session := server.NewSession(cli1, cli2)
		session.Round.Selection2 = game.SelectionRock
		if err := session.Select(cli1, game.SelectionRock); err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
		if session.Round.Selection1 != game.SelectionNone {
			t.Fatalf("Expected round selection1 to be none, but was %q!", session.Round.Selection1)
		}
		if session.Round.Selection2 != game.SelectionNone {
			t.Fatalf("Expected round selection2 to be none, but was %q!", session.Round.Selection2)
		}
	})
	t.Run("ReturnNilAfterNonDrawResult", func(t *testing.T) {
		t.Parallel()
		cli1 := server.NewClient(new(connMock))
		cli2 := server.NewClient(new(connMock))
		session := server.NewSession(cli1, cli2)
		session.Round.Selection2 = game.SelectionPaper
		if err := session.Select(cli1, game.SelectionRock); err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
		if session.Round.Selection1 != game.SelectionRock {
			t.Fatalf("Expected selection1 to be rock, but was %q!", session.Round.Selection1)
		}
		if session.Round.Selection2 != game.SelectionPaper {
			t.Fatalf("Expected selection2 to be rock, but was %q!", session.Round.Selection2)
		}
	})
}

func TestSessionClose(t *testing.T) {
	t.Parallel()
	cli1 := server.NewClient(new(connMock))
	cli2 := server.NewClient(new(connMock))
	session := server.NewSession(cli1, cli2)
	session.Close()
	if cli1.Session != nil {
		t.Fatalf("Expected cli1 session to be nil, but was %v!", cli1.Session)
	}
	if cli2.Session != nil {
		t.Fatalf("Expected cli2 session to be nil, but was %v!", cli2.Session)
	}
}
