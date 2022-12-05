package com

import (
	"encoding/json"
	"fmt"
	"io"
)

const bufferSize = 128

// WriteMessage marshals the given message into a JSON and writes it with the given writer.
func WriteMessage[T any](writer io.Writer, messageType MessageType, content T) error {
	bytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal %s content into JSON. %w", messageType, err)
	}
	if err := Write(writer, Message{Type: messageType, Content: bytes}); err != nil {
		return fmt.Errorf("failed to write %s message. %w", messageType, err)
	}
	return nil
}

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

// ReadMessage reads message from the reader and unmarshals it as a JSON data.
func ReadMessage[T any](reader io.Reader) (*T, error) {
	message, err := Read[Message](reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read %v message. %w", message, err)
	}
	content := new(T)
	if err := json.Unmarshal(message.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message content from JSON. %w", err)
	}
	return content, nil
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
