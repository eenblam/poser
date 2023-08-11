package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrRoomFull = errors.New("Room is full")

type Room struct {
	// This could also be a sync.Map,
	// but I don't think this use case fits what that's optimized for.
	mux    sync.Mutex
	Server *Server
	ID     string
	// Map to check membership of conn, as well as count of active players
	Conns map[*Connection]bool
	// Maximum number of players in room
	Size int
	// Ordered mapping of position to player
	Slots []*Connection
}

func NewRoom(id string, size int) *Room {
	//TODO validate room size, return error
	slots := make([]*Connection, size)
	for i := range slots {
		slots[i] = nil
	}
	return &Room{
		ID:    id,
		Conns: make(map[*Connection]bool),
		Size:  size,
		Slots: slots,
	}
}

func (r *Room) Add(conn *Connection) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if len(r.Conns) < r.Size {
		// Find a slot for user
		for i, slot := range r.Slots {
			if slot == nil { // add user
				r.Slots[i] = conn
				conn.PlayerNumber = i + 1
				break
			} else if i == (r.Size - 1) { // no slots!
				return fmt.Errorf("expected open slot in room %s but found none", r.ID)
			}
		}
		r.Conns[conn] = true
		return nil
	} else {
		// Feature: add user to queue, allow spectating
		log.Printf("room %s full at %d/%d", r.ID, len(r.Conns), r.Size)
		return ErrRoomFull
	}
}

// Delete conn from room, and return number of remaining connections.
//
// This allows for an atomic check of the length after deletion
// to confirm room is empty.
func (r *Room) Remove(conn *Connection) int {
	r.mux.Lock()
	defer r.mux.Unlock()
	delete(r.Conns, conn)
	// Remember: PlayerNumber is 1-indexed
	if conn.PlayerNumber <= len(r.Slots) {
		r.Slots[conn.PlayerNumber-1] = nil
	} else {
		log.Printf("Error: conn %s has player number %d in room of size %d", conn.ID, conn.PlayerNumber, r.Size)
	}
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
	bs, err := MakeMessage[T](messageType, message)
	if err != nil {
		return err
	}
	go room.Broadcast(from, bs)
	return nil
}

func (r *Room) BroadcastConnections() {
	r.mux.Lock()
	defer r.mux.Unlock()

	// Build list of IDs
	ids := make([]string, 0)
	// Get these from slots in order to maintain player order
	for _, conn := range r.Slots {
		if conn == nil {
			ids = append(ids, "")
		} else {
			ids = append(ids, conn.ID)
		}
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
