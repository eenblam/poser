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

// TODO make this configurable on room creation
const roomSize = 8

// Connection is a wrapper around websocket.Conn that also stores a local ID
type Connection struct {
	*websocket.Conn
	ID           string
	PlayerNumber int
}

func (c *Connection) Notify(message string, isErr bool) {
	bs, err := MakeMessage("notification", &NotificationMessage{
		Timestamp: time.Now(),
		Message:   message,
		IsError:   isErr,
	})
	if err != nil {
		log.Printf("failed to notify client: %s", err)
		return
	}
	c.WriteMessage(websocket.TextMessage, bs)
}

func (c *Connection) SendState(state State) {
	bs, err := MakeMessage("state", &StateMessage{State: state})
	if err != nil {
		log.Printf("failed to send state to client client: %s", err)
		return
	}
	c.WriteMessage(websocket.TextMessage, bs)
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
	//TODO get roomSize from / form
	room, loaded := s.RoomCache.LoadOrStore(roomId, NewRoom(roomId, roomSize))
	if !loaded {
		log.Printf("Created new room %s", roomId)
	}
	return room.(*Room)
}

func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	roomId, ok := mux.Vars(r)["room"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	room := s.GetOrCreateRoom(roomId)

	wsConn, err := upgrader.Upgrade(w, r, nil)
	conn := &Connection{
		Conn: wsConn,
		ID:   fmt.Sprintf("user-%s", uuid.New().String()),
		// PlayerNumber is for distinguishing things like user's color. Assigned later.
		PlayerNumber: 0,
	}
	defer conn.Close()
	if err != nil {
		// Upgrade() already wrote an error message, so just log error and return.
		log.Printf("failed to upgrade connection: %s", err)
		return
	}

	//TODO handle case where user ID already exists
	if err = room.Add(conn); err != nil {
		if err == ErrRoomFull {
			conn.Notify("Room is currently full.", true)
		} else {
			log.Printf("Error adding user to room: %s", err)
			conn.Notify("Couldn't add user to room.", true)
		}
		// For now, let's just close the websocket. Later, we could implement spectating.
		return
	}
	log.Printf("New connection from %s", conn.RemoteAddr())
	defer func() {
		//TODO handle room owner leaving (assign new?)
		//TODO handle prompt writer leaving (assign new prompt writer)
		//TODO handle final voter leaving (register voting as complete)
		//TODO handle fake artist leaving during game (make a note? or just end game?)
		//     Could just check for this after voting. (State PoserGuessing
		//     "You voted for X. Fake artist was Y, who left."
		//TODO handle fake artist leaving during guess (end game)
		log.Printf("Closing connection to %s", conn.RemoteAddr())
		if room.Remove(conn) == 0 { // If everyone has now left, delete the room
			log.Printf("Deleting room %s", room.ID)
			s.RoomCache.Delete(room.ID)
		} else { // Otherwise, let remaining users know this user left
			room.BroadcastConnections()
		}
	}()

	// Send user their ID
	bs, err := MakeMessage("connection", &ConnectionMessage{
		ID:           conn.ID,
		PlayerNumber: conn.PlayerNumber,
	})
	if err != nil {
		log.Printf("Error marshalling connection: %s", err)
		conn.WriteControl(websocket.CloseMessage,
			// Don't share actual error to avoid violating same-origin policy
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal error"),
			time.Now().Add(500*time.Millisecond))
		return
	}
	conn.WriteMessage(websocket.TextMessage, bs)
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
			m.PlayerNumber = conn.PlayerNumber
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
		case "start":
			// Nothing to parse from data
			if conn.PlayerNumber != 1 {
				conn.Notify(fmt.Sprintf("You cannot start the game as player #%d.", conn.PlayerNumber), true)
				continue LOOP
			}
			if room.State != Waiting {
				// Just ignore, player may have accidentally sent this
				continue LOOP
			}
			go room.Start()
		default:
			//DEBUG Just broadcast messages to all other room members for now
			log.Printf("%s:%s: unexpected message: %s", room.ID, conn.RemoteAddr(), message)
			go room.Broadcast(conn, message)
		}

	}
}
