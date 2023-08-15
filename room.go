package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrRoomFull = errors.New("room is full")
var ErrGameInProgress = errors.New("game is in progress")

type Role string

const (
	Artist Role = "Artist"
	Muse   Role = "Muse"
	Poser  Role = "Poser"
)

// A Room is a lobby of players, and includes any Game the players start within the lobby.
//
// Room includes a few kinds of methods:
// * Adding and removing clients as they connect and disconnect
// * Communication methods for sending data to clients
// * Wrappers around Game transition functions to handle I/O relevant to the transition. These methods share the same names as the wrapped methods.
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
	// Game state machine
	Game *Game
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
		Game:  &Game{State: Waiting},
	}
}

func (r *Room) Add(conn *Connection) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if !r.Game.IsJoinable() {
		return ErrGameInProgress
	}
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

/* Room communication methods */

func (r *Room) Broadcast(from *Connection, message []byte) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.broadcastUnsafe(from, message)
}

// BroadcastType broadcasts a message of type T to all connections in the room.
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

// BroadcastConnections informs all clients in the room of the current list of players.
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
	bs, err := MakeMessage[PlayersMessage]("players", PlayersMessage{IDs: ids})
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

/* Game state methods */

func (r *Room) SetPrompt(prompt string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	err := r.Game.SetPrompt(prompt)
	if err != nil {
		log.Printf("error setting prompt: %s", err)
		r.abortGameUnsafe(fmt.Sprintf("Couldn't set prompt: %s", err))
		return
	}
	r.broadcastStateUnsafe()

	// Notify everyone but Poser of prompt
	poser := r.Slots[r.Game.Poser]
	poser.Notify(
		"You are the poser! Just act cool, play along, and try to guess what you're drawing.",
		false,
	)
	for c := range r.Conns {
		if c != poser {
			c.Notify(fmt.Sprintf("The prompt is: %s", prompt), false)
		}
	}
}

func (r *Room) Start() {
	r.mux.Lock()
	defer r.mux.Unlock()

	players := r.getActivePlayerNumbers()

	err := r.Game.Start(players)
	if err == ErrGameInProgress {
		// Here, this could mean the client just clicked a few times, so we can disregard.
		log.Println("error: Start() issued for in-progress game")
		return
	} else if err != nil {
		log.Printf("error starting game: %s", err)
		r.abortGameUnsafe("Couldn't start game. Not enough players.")
		return
	}

	log.Printf("Starting game for room %s", r.ID)
	// Notify all, but don't reveal the Muse to other players here!
	// Doing so reduces the number of possible fake artists, which is less fun in small games.
	r.notifyAllUnsafe("Game starting! The Muse is contemplating...", false)
	r.broadcastStateUnsafe()

	// Notify Muse
	muse := r.Slots[r.Game.Muse]
	muse.Notify("You are the Muse! Pick a prompt for the round.", false)
	// Send role to Muse
	bs, err := MakeMessage("role", &RoleMessage{Role: Muse})
	if err != nil {
		log.Printf("error marshalling role message: %s", err)
		r.abortGameUnsafe("Whoops! There was an error starting the game.")
		return
	}
	err = muse.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		log.Printf("error sending role message: %s", err)
		r.abortGameUnsafe("Whoops! There was an error starting the game.")
		return
	}
}

// getActivePlayerNumbers returns a slice of player numbers.
// This provides the list of indices, skipping any nils.
// So if four players join, and the second leaves, this returns [0, 2, 3].
//
// Not threadsafe.
func (r *Room) getActivePlayerNumbers() []int {
	choices := make([]int, 0, len(r.Conns))
	for i, conn := range r.Slots {
		if conn != nil {
			choices = append(choices, i)
		}
	}
	return choices
}

// abortGameUnsafe resets game state, sends error to all clients, and updates state for UI.
// Not threadsafe.
func (r *Room) abortGameUnsafe(message string) {
	r.Game.Abort()
	r.notifyAllUnsafe(message, true)
	r.broadcastStateUnsafe()
}

// notifyAllUnsafe sends a Notification to all clients.
// Not threadsafe.
func (r *Room) notifyAllUnsafe(message string, isErr bool) {
	for conn := range r.Conns {
		conn.Notify(message, isErr)
	}
}

// broadcastStateUnsafe sends current game state to all clients.
// Not threadsafe.
func (r *Room) broadcastStateUnsafe() {
	bs, err := MakeMessage("state", &StateMessage{State: r.Game.State})
	if err != nil {
		log.Printf("failed to create StateMessage for broadcast")
		return
	}
	r.broadcastUnsafe(nil, bs)
}
