package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func MustRead(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read file %s: %s", file, err)
	}
	return string(b)
}

var indexHTML = MustRead("./frontend/dist/home.html")
var roomHTML = MustRead("./frontend/dist/app.html")

// The home page serves a page with a button to create a new room
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(indexHTML))
	case http.MethodPost:
		NewRoomHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// NewRoomHandler just creates a UUID for a new room, then redirects the user.
//
// It would be cool to store all recent room IDs in a cookie,
// then render the list on the homepage for a user to return to.
// But for now we just forward them to a room URL, which will autogenerate the room.
func NewRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Homepage requested a new room
	//TODO could instead do some sort of TinyURL style Base58(SHA256(url, username))

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := uuid.New()
	roomPath := fmt.Sprintf("/room/%s", id)
	http.Redirect(w, r, roomPath, http.StatusFound)
}

// RoomHandler serves the room assets
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte(roomHTML))
}
