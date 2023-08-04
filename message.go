package main

import (
	"encoding/json"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ChatMessage struct {
	ID string `json:"id"`
	//TODO sanitize this
	Text string `json:"text"`
	//TODO can I parse a timestamp as a time.Time?
	Timestamp int    `json:"timestamp"`
	User      string `json:"user"`
}

type DrawMessage struct {
	LastX int `json:"lastX"`
	LastY int `json:"lastY"`
	X     int `json:"x"`
	Y     int `json:"y"`
}

func ParseMessage(bs []byte) (messageType string, data []byte, err error) {
	var message Message
	err = json.Unmarshal(bs, &message)
	if err != nil {
		return "", nil, err
	}
	return message.Type, message.Data, nil
}
