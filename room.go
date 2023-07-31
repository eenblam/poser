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
	for conn := range r.Conns {
		if conn != nil && conn != from {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// TODO this could be a LOT nicer
func (r *Room) BroadcastConnections() {
	r.mux.Lock()
	defer r.mux.Unlock()
	ids := make([]string, 0)
	for conn := range r.Conns {
		ids = append(ids, conn.ID)
	}
	bs, err := json.Marshal(struct {
		Type string   `json:"type"`
		IDs  []string `json:"ids"`
	}{Type: "ids", IDs: ids})
	if err != nil {
		log.Printf("Error marshalling connections: %s", err)
		return
	}
	for conn := range r.Conns {
		if conn != nil {
			conn.WriteMessage(websocket.TextMessage, bs)
		}
	}
}
