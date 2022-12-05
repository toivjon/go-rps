package com_test

import (
	"errors"
	"testing"

	"github.com/toivjon/go-rps/internal/com"
)

var errMock = errors.New("mock-error")

type writerMock struct {
	n   int
	err error
}

func (w *writerMock) Write(b []byte) (int, error) {
	return w.n, w.err
}

func TestWriteMessage(t *testing.T) {
	t.Parallel()
	t.Run("ReturnsErrorWhenMarshallingFails", func(t *testing.T) {
		t.Parallel()
		if err := com.WriteMessage(nil, com.TypeJoin, make(chan int)); err == nil {
			t.Fatalf("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsErrorWhenWriterWriteFails", func(t *testing.T) {
		t.Parallel()
		writer := &writerMock{n: 0, err: errMock}
		if err := com.WriteMessage(writer, com.TypeJoin, com.JoinContent{Name: ""}); err == nil {
			t.Fatalf("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsNilWhenSuccess", func(t *testing.T) {
		t.Parallel()
		writer := &writerMock{n: 0, err: nil}
		if err := com.WriteMessage(writer, com.TypeJoin, com.JoinContent{Name: ""}); err != nil {
			t.Fatalf("Expected no error, but error was returned: %s", err)
		}
	})
}

func TestWrite(t *testing.T) {
	t.Parallel()
	t.Run("ReturnsErrorWhenMarshallingFails", func(t *testing.T) {
		t.Parallel()
		if err := com.Write(nil, make(chan int)); err == nil {
			t.Fatalf("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsErrorWhenWriterWriteFails", func(t *testing.T) {
		t.Parallel()
		writer := &writerMock{n: 0, err: errMock}
		if err := com.Write(writer, ""); err == nil {
			t.Fatalf("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsNilWhenSuccess", func(t *testing.T) {
		t.Parallel()
		writer := &writerMock{n: 0, err: nil}
		if err := com.Write(writer, ""); err != nil {
			t.Fatalf("Expected no error, but error was returned: %s", err)
		}
	})
}

type readerMock struct {
	n   int
	err error
	val []byte
}

func (r *readerMock) Read(b []byte) (int, error) {
	copy(b, r.val)
	return r.n, r.err
}

func TestReadMessage(t *testing.T) {
	t.Parallel()
	t.Run("ReturnsErrorWhenReaderReadFails", func(t *testing.T) {
		t.Parallel()
		reader := &readerMock{n: 0, err: errMock, val: nil}
		if _, err := com.ReadMessage[com.JoinContent](reader); err == nil {
			t.Fatal("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsErrorWhenUnmarshallingFails", func(t *testing.T) {
		t.Parallel()
		data := `{"type":"JOIN","content":{invalid}}`
		reader := &readerMock{n: len(data), err: nil, val: []byte(data)}
		if _, err := com.ReadMessage[com.JoinContent](reader); err == nil {
			t.Fatal("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsResultWhenSuccess", func(t *testing.T) {
		t.Parallel()
		expectedResult := struct{}{}
		data := `{"type":"JOIN","content":{"name":"donald"}}`
		reader := &readerMock{n: len(data), err: nil, val: []byte(data)}
		val, err := com.ReadMessage[struct{}](reader)
		if err != nil {
			t.Fatalf("Expected no error, but error was returned: %s", err)
		}
		if *val != expectedResult {
			t.Fatalf("Expected %s but received %s", expectedResult, *val)
		}
	})
}

func TestRead(t *testing.T) {
	t.Parallel()
	t.Run("ReturnsErrorWhenReaderReadFails", func(t *testing.T) {
		t.Parallel()
		reader := &readerMock{n: 0, err: errMock, val: nil}
		if _, err := com.Read[string](reader); err == nil {
			t.Fatal("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsErrorWhenUnmarshallingFails", func(t *testing.T) {
		t.Parallel()
		reader := &readerMock{n: 0, err: nil, val: nil}
		if _, err := com.Read[chan int](reader); err == nil {
			t.Fatal("Expected an error, but nil was returned!")
		}
	})
	t.Run("ReturnsResultWhenSuccess", func(t *testing.T) {
		t.Parallel()
		expectedResult := struct{}{}
		reader := &readerMock{n: 2, err: nil, val: []byte("{}")}
		val, err := com.Read[struct{}](reader)
		if err != nil {
			t.Fatalf("Expected no error, but error was returned: %s", err)
		}
		if *val != expectedResult {
			t.Fatalf("Expected %s but received %s", expectedResult, *val)
		}
	})
}
