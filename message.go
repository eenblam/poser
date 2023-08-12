package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ChatMessage struct {
	ID           string `json:"id"`
	PlayerNumber int    `json:"playerNumber"`
	//TODO sanitize this
	Text string `json:"text"`
	//TODO can I parse a timestamp as a time.Time?
	Timestamp int    `json:"timestamp"`
	User      string `json:"user"`
}

type ConnectionMessage struct {
	ID           string `json:"id"`
	PlayerNumber int    `json:"playerNumber"`
}

type DrawMessage struct {
	LastX        int `json:"lastX"`
	LastY        int `json:"lastY"`
	X            int `json:"x"`
	Y            int `json:"y"`
	PlayerNumber int `json:"playerNumber"`
}

type NotificationMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	IsError   bool      `json:"isError"`
}

func ParseMessage(bs []byte) (messageType string, data []byte, err error) {
	var message Message
	err = json.Unmarshal(bs, &message)
	if err != nil {
		return "", nil, err
	}
	return message.Type, message.Data, nil
}

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
