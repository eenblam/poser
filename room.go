package main

import (
	"encoding/json"
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

func (r *Room) Remove(conn *Connection) {
	r.mux.Lock()
	defer r.mux.Unlock()
	delete(r.Conns, conn)
}

func (r *Room) String() string {
	return r.ID
}

func (r *Room) Broadcast(from *Connection, message []byte) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.broadcastUnsafe(from, message)
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
