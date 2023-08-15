package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Message provides a wrapper struct for parsing and sending messages to the client.
//
// It's really a union type, but Go doesn't have those,
// so we're using a struct with a type annotation and a raw JSON field.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ChatMessage ferries chat info between users.
type ChatMessage struct {
	ID           string `json:"id"`
	PlayerNumber int    `json:"playerNumber"`
	//TODO sanitize this
	Text string `json:"text"`
	//TODO can I parse a timestamp as a time.Time?
	Timestamp int    `json:"timestamp"`
	User      string `json:"user"`
}

// ConnectionMessage is used to welcome a new user to the game,
// providing the most basic info needed by the client: ID and player number.
//
// ID and number aren't equivalent, since player N could leave the room and
// be replaced by someone else with a different ID.
type ConnectionMessage struct {
	ID           string `json:"id"`
	PlayerNumber int    `json:"playerNumber"`
}

// DrawMessage simply forwards the coordinates of a draw event to the client.
//
// Draw events are broken up into single strokes of a larger vector.
type DrawMessage struct {
	LastX        int `json:"lastX"`
	LastY        int `json:"lastY"`
	X            int `json:"x"`
	Y            int `json:"y"`
	PlayerNumber int `json:"playerNumber"`
}

// PlayersMessage notifies a client of the current players in the game.
type PlayersMessage struct {
	IDs []string `json:"ids"`
}

// PromptMessage, sent by the Muse to the server, contains the Muse's prompt.
type PromptMessage struct {
	Prompt string `json:"prompt"`
}

// RoleMessage notifies a client of its role in the game.
type RoleMessage struct {
	Role Role `json:"role"`
}

// StateMessage notifies a client of the current state of the game.
type StateMessage struct {
	State State `json:"state"`
}

// NotificationMessage is used to provide messages from the server to the client.
//
// These could potentially be consumed by chat instead of a separate notification widget;
// that's up to the client to decide.
type NotificationMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	IsError   bool      `json:"isError"`
}

// ParseMessage unwraps a Message from the client, but doesn't try to parse the inner data.
func ParseMessage(bs []byte) (messageType string, data []byte, err error) {
	var message Message
	err = json.Unmarshal(bs, &message)
	if err != nil {
		return "", nil, err
	}
	return message.Type, message.Data, nil
}

// MakeMessage wraps a message in a Message struct, then marshals to JSON bytes
// to be sent to the client. In theory, a message can be any type T.
func MakeMessage[T any](messageType string, message T) ([]byte, error) {
	rawJson, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("error marshalling message of type %T to data: %w", message, err)
	}
	bs, err := json.Marshal(&Message{Type: messageType, Data: rawJson})
	if err != nil {
		return nil, fmt.Errorf("error marshalling message: %w", err)
	}
	return bs, nil
}
