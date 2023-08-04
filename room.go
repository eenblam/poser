package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	// This could also be a sync.Map,
	// but I don't think this use case fits what that's optimized for.
	mux    sync.Mutex
	Server *Server
	ID     string
	Conns  map[*Connection]bool
}

func NewRoom(id string) *Room {
	return &Room{ID: id, Conns: make(map[*Connection]bool)}
}

func (r *Room) Add(conn *Connection) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.Conns[conn] = true
}

// Delete conn from room, and return number of remaining connections.
//
// This allows for an atomic check of the length after deletion
// to confirm room is empty.
func (r *Room) Remove(conn *Connection) int {
	r.mux.Lock()
	defer r.mux.Unlock()
	delete(r.Conns, conn)
	return len(r.Conns)
}

func (r *Room) String() string {
	return r.ID
}

func (r *Room) Broadcast(from *Connection, message []byte) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.broadcastUnsafe(from, message)
}

// BroadcastT broadcasts a message of type T to all connections in the room.
// If from is non-nil, that connection will be omitted.
// If the message is successfully marshalled to JSON, it will be sent in a separate goroutine.
//
// Note that this can't be a method since it's a generic function.
func BroadcastType[T any](room *Room, from *Connection, messageType string, message T) error {
	rawJson, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message of type %T to data: %w", message, err)
	}
	bs, err := json.Marshal(&Message{Type: messageType, Data: rawJson})
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}
	go room.Broadcast(from, bs)
	return nil
}

func (r *Room) BroadcastConnections() {
	r.mux.Lock()
	defer r.mux.Unlock()

	// Build list of IDs
	ids := make([]string, 0)
	for conn := range r.Conns {
		ids = append(ids, conn.ID)
	}
	// Create JSON
	bs, err := json.Marshal(struct {
		Type string   `json:"type"`
		IDs  []string `json:"ids"`
	}{Type: "ids", IDs: ids})
	if err != nil {
		log.Printf("Error marshalling connections: %s", err)
		return
	}
	// Broadcast to *all* connections (hence from=nil)
	r.broadcastUnsafe(nil, bs)
}

// Non-threadsafe broadcast; callers must handle locking.
func (r *Room) broadcastUnsafe(from *Connection, message []byte) {
	for conn := range r.Conns {
		if conn != nil && conn != from {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
