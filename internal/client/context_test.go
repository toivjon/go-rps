package client_test

import (
	"testing"

	"github.com/toivjon/go-rps/internal/client"
)

type readerMock struct {
	n   int
	err error
	val []byte
}

func failingReaderMock(err error) readerMock {
	return readerMock{
		n:   0,
		err: err,
		val: nil,
	}
}

func succeedingReaderMock(data string) readerMock {
	data += "\n"
	return readerMock{
		n:   len(data),
		err: nil,
		val: []byte(data),
	}
}

func (r readerMock) Read(b []byte) (int, error) {
	copy(b, r.val)
	return r.n, r.err
}

type writerMock struct {
	n   int
	err error
}

func (r writerMock) Write(b []byte) (int, error) {
	return r.n, r.err
}

type readWriterMock struct {
	readerMock
	writerMock
}

func TestNewContext(t *testing.T) {
	t.Parallel()
	input := new(readerMock)
	conn := new(readWriterMock)
	ctx := client.NewContext(input, conn)
	if ctx.Input != input {
		t.Fatalf("Expected to contain input member %#p but had %#p", input, &ctx.Input)
	}
	if ctx.Conn != conn {
		t.Fatalf("Expected to contain conn member %#p but had %#p", conn, &ctx.Conn)
	}
}
