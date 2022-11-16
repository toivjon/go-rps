package com

import (
	"encoding/json"
	"fmt"
	"io"
)

const bufferSize = 128

// Write marshals the provided data into a JSON and writes it with the given writer.
func Write[T any](writer io.Writer, data T) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data into JSON. %w", err)
	}
	if _, err = writer.Write(bytes); err != nil {
		return fmt.Errorf("failed to write data into connection. %w", err)
	}
	return nil
}

// WriteConnect writes a CONNECT message as a JSON with the given writer.
func WriteConnect(writer io.Writer, content ConnectContent) error {
	bytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal CONNECT content: %w", err)
	}
	out := Message{Type: MessageTypeConnect, Content: bytes}
	if err := Write(writer, out); err != nil {
		return fmt.Errorf("failed to write CONNECT message: %w", err)
	}
	return nil
}

// WriteConnected writes a CONNECTED message as a JSON with the given writer.
func WriteConnected(writer io.Writer) error {
	out := Message{Type: MessageTypeConnected, Content: nil}
	if err := Write(writer, out); err != nil {
		return fmt.Errorf("failed to write CONNECTED message: %w", err)
	}
	return nil
}

// Read reads data from the reader and unmarshals it as a JSON data.
func Read[T any](reader io.Reader) (*T, error) {
	buffer := make([]byte, bufferSize)
	byteCount, err := reader.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from connection. %w", err)
	}
	out := new(T)
	if err := json.Unmarshal(buffer[:byteCount], out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data from JSON. %w", err)
	}
	return out, nil
}
