package client_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/toivjon/go-rps/internal/client"
	"github.com/toivjon/go-rps/internal/game"
)

var errMock = errors.New("mock-error")

func TestRun(t *testing.T) {
	t.Parallel()
	ctx := client.NewContext(new(readerMock), new(readWriterMock))
	t.Run("ReturnErrorWhenStateFails", func(t *testing.T) {
		t.Parallel()
		state := func(client.Context) (client.State, error) { return nil, errMock }
		if err := client.Run(ctx, state); !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnNilWhenSuccessAfterOneIteration", func(t *testing.T) {
		t.Parallel()
		state := func(client.Context) (client.State, error) { return nil, client.ErrEnd }
		if err := client.Run(ctx, state); err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
	t.Run("ReturnNilWhenSuccessAfterTwoIterations", func(t *testing.T) {
		t.Parallel()
		state1 := func(client.Context) (client.State, error) { return nil, client.ErrEnd }
		state2 := func(client.Context) (client.State, error) { return state1, nil }
		if err := client.Run(ctx, state2); err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}

func TestConnected(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenInputScanningFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(failingReaderMock(errMock), new(readWriterMock))
		result, err := client.Connected(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %s in the chain, but did not exists!", errMock)
		}
	})
	t.Run("ReturnErrorWhenInputValidationFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(succeedingReaderMock(strings.Repeat("s", client.NameMaxLength+1)), new(readWriterMock))
		result, err := client.Connected(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, client.ErrNameTooLong) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", client.ErrNameTooLong, err)
		}
	})
	t.Run("ReturnErrorWhenWriteMessageFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(
			succeedingReaderMock(strings.Repeat("s", client.NameMaxLength)),
			readWriterMock{
				readerMock: readerMock{n: 0, err: nil, val: nil},
				writerMock: writerMock{n: 0, err: errMock},
			})
		result, err := client.Connected(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnStateWhenSuccess", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(
			succeedingReaderMock(strings.Repeat("s", client.NameMaxLength)),
			readWriterMock{
				readerMock: readerMock{n: 0, err: nil, val: nil},
				writerMock: writerMock{n: 0, err: nil},
			})
		result, err := client.Connected(ctx)
		if result == nil {
			t.Fatalf("Expected non-nil result, but nil was returned!")
		}
		if err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}

func TestJoined(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenReadFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(new(readerMock), newReadableConnMock("", errMock))
		result, err := client.Joined(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnErrorWhenUnmarshalFails", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"JOIN","content":"non-json"}`
		ctx := client.NewContext(new(readerMock), newReadableConnMock(data, nil))
		result, err := client.Joined(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if err == nil {
			t.Fatal("Expected non-nil error, but nil was returned!")
		}
	})
	t.Run("ReturnStateWhenSuccess", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"JOIN","content":{"name":"donald"}}`
		ctx := client.NewContext(new(readerMock), newReadableConnMock(data, nil))
		result, err := client.Joined(ctx)
		if result == nil {
			t.Fatal("Expected non-nil result, but nil was returned!")
		}
		if err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
	})
}

func TestStarted(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenInputScanningFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(failingReaderMock(errMock), new(readWriterMock))
		result, err := client.Started(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %s in the chain, but did not exists!", errMock)
		}
	})
	t.Run("ReturnErrorWhenInputValidationFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(succeedingReaderMock("x"), new(readWriterMock))
		result, err := client.Started(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if err == nil {
			t.Fatal("Expected non-nil error, but nil was returned!")
		}
	})
	t.Run("ReturnErrorWhenWriteMessageFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(
			succeedingReaderMock(string(game.SelectionPaper)),
			readWriterMock{
				readerMock: readerMock{n: 0, err: nil, val: nil},
				writerMock: writerMock{n: 0, err: errMock},
			})
		result, err := client.Started(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnStateWhenSuccess", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(
			succeedingReaderMock(string(game.SelectionPaper)),
			readWriterMock{
				readerMock: readerMock{n: 0, err: nil, val: nil},
				writerMock: writerMock{n: 0, err: nil},
			})
		result, err := client.Started(ctx)
		if result == nil {
			t.Fatalf("Expected non-nil result, but nil was returned!")
		}
		if err != nil {
			t.Fatalf("Expected nil error, but %q was returned!", err)
		}
	})
}

func TestWaiting(t *testing.T) {
	t.Parallel()
	t.Run("ReturnErrorWhenReadFails", func(t *testing.T) {
		t.Parallel()
		ctx := client.NewContext(new(readerMock), newReadableConnMock("", errMock))
		result, err := client.Waiting(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, errMock) {
			t.Fatalf("Expected %q error in the chain %q, but did not exists!", errMock, err)
		}
	})
	t.Run("ReturnErrorWhenUnmarshalFails", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"RESULT","content":"non-json"}`
		ctx := client.NewContext(new(readerMock), newReadableConnMock(data, nil))
		result, err := client.Waiting(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if err == nil {
			t.Fatal("Expected non-nil error, but nil was returned!")
		}
	})
	t.Run("ReturnStateWhenSuccessWithDraw", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"RESULT","content":{"opponentSelection":"s","result":"DRAW"}}`
		ctx := client.NewContext(new(readerMock), newReadableConnMock(data, nil))
		result, err := client.Waiting(ctx)
		if result == nil {
			t.Fatal("Expected non-nil result, but nil was returned!")
		}
		if err != nil {
			t.Fatalf("Expected no error, but %q was returned!", err)
		}
	})
	t.Run("ReturnNilWhenSuccessWithNoDraw", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"RESULT","content":{"opponentSelection":"s","result":"WIN"}}`
		ctx := client.NewContext(new(readerMock), newReadableConnMock(data, nil))
		result, err := client.Waiting(ctx)
		if result != nil {
			t.Fatalf("Expected nil result, but %v was returned!", result)
		}
		if !errors.Is(err, client.ErrEnd) {
			t.Fatalf("Expected %q error, but %q was returned!", client.ErrEnd, err)
		}
	})
}

func newReadableConnMock(data string, err error) readWriterMock {
	data += "\n"
	return readWriterMock{
		readerMock: readerMock{
			n:   len(data),
			err: err,
			val: []byte(data),
		}, writerMock: writerMock{
			n:   0,
			err: nil,
		},
	}
}
