package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrRoomFull = errors.New("room is full")
var ErrGameInProgress = errors.New("game is in progress")

// Game states
type State string

const (
	Waiting       State = "Waiting"
	GettingPrompt State = "GettingPrompt"
	Drawing       State = "Drawing"
	Voting        State = "Voting"
	PoserGuessing State = "PoserGuessing"
	// ValidatingGuess?
)

// Player roles
type Role string

const (
	Artist Role = "Artist"
	Muse   Role = "Muse"
	Poser  Role = "Poser"
)

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
	// Current state of room's game
	State State
	// Muse for current game
	Muse *Connection
	// Fake Artist for current game
	Poser *Connection
	// PlayerNumber for first player
	FirstPlayerNumber int
	// Prompt for the game
	Prompt string
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
		State: Waiting,
	}
}

func (r *Room) Add(conn *Connection) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if !r.IsJoinable() {
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

// IsJoinable checks if room can currently be joined by a user. Not thread-safe.
func (r *Room) IsJoinable() bool {
	return r.State == Waiting
}

func (r *Room) SetPrompt(prompt string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.Prompt = prompt

	var err error
	// Pick Poser
	r.Poser, err = r.pickPoser()
	if err != nil {
		log.Printf("error picking poser: %s", err)
		r.abortGameUnsafe("Couldn't pick poser. Not enough players.")
		return
	}
	// Notify everyone but Poser of prompt
	r.Poser.Notify(
		"You are the poser! Just act cool, play along, and try to guess what you're drawing.",
		false,
	)
	for c := range r.Conns {
		if c != r.Poser {
			c.Notify(fmt.Sprintf("The prompt is: %s", prompt), false)
		}
	}
	// Pick first player
	r.FirstPlayerNumber, err = r.pickFirstPlayer()
	if err != nil {
		log.Printf("error picking first player: %s", err)
		r.abortGameUnsafe("Couldn't pick first player. Not enough players.")
		return
	}
	// Update state
	r.State = Drawing
	r.broadcastStateUnsafe()
}

func (r *Room) Start() {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.State != Waiting { // Game already started
		// Could just be multiple sends from client:
		//   Client sends Start A
		//   Start() locks for A
		//   Client double clicked, sends Start() B. Waits for lock.
		//   Start A completes, unlocks.
		//   Start() B acquires lock. Sees state is not Waiting. Here we are!
		// So don't fail here, but do log the error. This indicates an issue with client UI.
		log.Println("error: Start() issued for in-progress game")
		return
	}
	log.Printf("Starting game for room %s", r.ID)
	r.State = GettingPrompt
	museIndex, err := r.pickMuse()
	if err != nil {
		log.Printf("error picking prompter: %s", err)
		r.abortGameUnsafe("Whoops! There was an error choosing a player to set the prompt.")
		return
	}
	museConn := r.Slots[museIndex]
	if museConn == nil { // big problem! we somehow picked a bad slot within a lock
		log.Println("error picking prompter: somehow picked an empty slot")
		r.abortGameUnsafe("Whoops! There was an error choosing a player to set the prompt.")
		return
	}
	r.Muse = museConn
	// Don't reveal the Muse to other players here!
	// Doing so reduces the number of possible fake artists, which is less fun in small games.
	r.notifyAllUnsafe("Game starting! The Muse is contemplating...", false)
	r.broadcastStateUnsafe()
	// Send prompter role
	bs, err := MakeMessage("role", &RoleMessage{Role: Muse})
	if err != nil {
		log.Printf("error marshalling start message: %s", err)
		r.abortGameUnsafe("Whoops! There was an error starting the game.")
		return
	}
	r.Muse.WriteMessage(websocket.TextMessage, bs)
	r.Muse.Notify("You are the Muse! Pick a prompt for the round.", false)
	//TODO handle fake artist similarly
}

// pickMuse selects one of the currently active clients to pick the prompt for the new round.
//
// Not threadsafe.
func (r *Room) pickMuse() (int, error) {
	//TODO randomly select someone, instead of just using owner
	playerIndex := 0
	return playerIndex, nil
}

// pickFirstPlayer selects an active clients to play first.
//
// Not threadsafe. Errors on empty list.
func (r *Room) pickFirstPlayer() (int, error) {
	if len(r.Conns) < 1 {
		return 0, errors.New("no players")
	}
	// We could try to be clever here and not allocate,
	// but this is only once per game. Not a hot path.
	// Grab indices (player numbers) of non-nil slots
	choices := make([]int, len(r.Conns))
	for i, conn := range r.Slots {
		if conn != nil {
			choices = append(choices, i)
		}
	}
	choiceIndex := rand.Intn(len(choices))
	return choices[choiceIndex], nil
}

// pickPoser selects an active clients to be the Poser. Doesn't pick the Muse.
//
// Not threadsafe.
// Errors if no one can be the Poser.
func (r *Room) pickPoser() (*Connection, error) {
	// We could try to be clever here and not allocate,
	// but this is only once per game. Not a hot path.
	nChoices := len(r.Conns) - 1
	if nChoices < 1 {
		// At least one player must be Muse, so this means no one can be the Poser.
		return nil, errors.New("no one available to to play the Poser")
	}
	choices := make([]*Connection, 0, nChoices)
	for conn := range r.Conns {
		if conn != r.Muse && conn != nil {
			choices = append(choices, conn)
		}
	}
	poserIndex := rand.Intn(nChoices)
	return choices[poserIndex], nil
}

// abortGameUnsafe resets game state, sends error to all clients, and updates state for UI.
// Not threadsafe.
func (r *Room) abortGameUnsafe(message string) {
	r.State = Waiting
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
	bs, err := MakeMessage("state", &StateMessage{State: r.State})
	if err != nil {
		log.Printf("failed to create StateMessage for broadcast")
		return
	}
	r.broadcastUnsafe(nil, bs)
}
