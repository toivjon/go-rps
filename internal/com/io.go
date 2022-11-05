package com

import (
	"encoding/json"
	"fmt"
)

const bufferSize = 128

// Writer is able to write the provided bytes of data.
type Writer interface {
	Write(b []byte) (n int, err error)
}

// Write marshals the provided data into a JSON and writes it with the given writer.
func Write[T any](writer Writer, data T) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data into JSON. %w", err)
	}
	if _, err = writer.Write(bytes); err != nil {
		return fmt.Errorf("failed to write data into connection. %w", err)
	}
	return nil
}

// Reader is able to read data from the target.
type Reader interface {
	Read(b []byte) (n int, err error)
}

// Read reads data from the reader and unmarshals it as a JSON data.
func Read[T any](reader Reader) (*T, error) {
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