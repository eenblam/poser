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
	conn := &Connection{wsConn, fmt.Sprintf("user-%s", uuid.New().String())}
	defer conn.Close()
	if err != nil {
		// Upgrade() already wrote an error message, so just log error and return.
		log.Printf("failed to upgrade connection: %s", err)
		return
	}

	//TODO limit room size. Error if room is full.
	//TODO handle case where user ID already exists
	room.Add(conn)
	log.Printf("New connection from %s", conn.RemoteAddr())
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

LOOP:
	for {
		// Break if we can't parse websocket message, continue if we can't parse app message
		mt, message, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			log.Printf("error reading message: %s", err)
			break
		}

		// basically the same processing for the parsed message as for the websocket message
		messageType, data, err := ParseMessage(message)
		if err != nil {
			log.Printf("Error parsing message: %s", err)
			continue LOOP
		}
		switch messageType {
		case "chat":
			m := &ChatMessage{}
			err := json.Unmarshal(data, m)
			if err != nil {
				log.Printf("Error unmarshalling chat message: %s", err)
				continue LOOP
			}
			//TODO check timestamp?
			//TODO sanitize text
			// Set user ID, ignore anything client may have set.
			m.User = conn.ID
			// Set message ID - these have to be distinct on the client side.
			m.ID = fmt.Sprintf("msg-%s", uuid.New().String())
			log.Printf("%s:%s: %s", room.ID, conn.RemoteAddr(), message)
			err = BroadcastType[*ChatMessage](room, nil, "chat", m)
			if err != nil {
				log.Println(err)
			}
			continue LOOP
		case "draw":
			m := &DrawMessage{}
			err := json.Unmarshal(data, m)
			if err != nil {
				log.Printf("Error unmarshalling draw message: %s", err)
				continue LOOP
			}
			//TODO set source user ID
			err = BroadcastType[*DrawMessage](room, conn, "draw", m)
			if err != nil {
				log.Println(err)
			}
			continue LOOP
		default:
			//DEBUG Just broadcast messages to all other room members for now
			log.Printf("%s:%s: unexpected message: %s", room.ID, conn.RemoteAddr(), message)
			go room.Broadcast(conn, message)
		}

	}
}

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
