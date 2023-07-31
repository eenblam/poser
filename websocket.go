package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Taken with much inspiration from https://dev.to/nyxtom/realtime-collaborative-drawing-with-canvas-and-webrtc-2d01

// Connection is a wrapper around websocket.Conn that also stores a local ID
type Connection struct {
	*websocket.Conn
	ID string
}

var upgrader = websocket.Upgrader{
	//DEBUG currently accepting all requests
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	RoomCache sync.Map
}

func NewServer() *Server {
	return &Server{RoomCache: sync.Map{}}
}

func (s *Server) GetOrCreateRoom(roomId string) *Room {
	room, loaded := s.RoomCache.LoadOrStore(roomId, NewRoom(roomId))
	if !loaded {
		log.Printf("Created new room %s", roomId)
	}
	return room.(*Room)
}

func (s *Server) Echo(w http.ResponseWriter, r *http.Request) {
	roomId, ok := mux.Vars(r)["room"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	room := s.GetOrCreateRoom(roomId)

	wsConn, err := upgrader.Upgrade(w, r, nil)
	conn := &Connection{wsConn, fmt.Sprintf("conn-%s", uuid.New().String())}
	defer conn.Close()
	if err != nil {
		// Upgrade() already wrote an error message, so just log error and return.
		log.Printf("failed to upgrade connection: %s", err)
		return
	}
	log.Printf("New connection from %s", conn.RemoteAddr())

	//TODO limit room size. Error if room is full.
	room.Add(conn)
	defer func() {
		log.Printf("Closing connection to %s", conn.RemoteAddr())
		if room.Remove(conn) == 0 { // If everyone has now left, delete the room
			log.Printf("Deleting room %s", room.ID)
			s.RoomCache.Delete(room.ID)
		} else { // Otherwise, let remaining users know this user left
			room.BroadcastConnections()
		}
	}()

	// Send user their ID
	data, err := json.Marshal(struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{Type: "connection", ID: conn.ID})
	if err != nil {
		log.Printf("Error marshalling connection: %s", err)
		conn.WriteControl(websocket.CloseMessage,
			// Don't share actual error to avoid violating same-origin policy
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal error"),
			time.Now().Add(500*time.Millisecond))
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
	// Send all IDs
	room.BroadcastConnections()

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			log.Printf("error reading message: %s", err)
			break
		}

		// Just broadcast messages to all other room members for now
		log.Printf("%s:%s: %s", room.ID, conn.RemoteAddr(), message)
		go room.Broadcast(conn, message)
	}
}
