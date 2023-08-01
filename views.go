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

var roomHTML = MustRead("./frontend/dist/index.html")

// The home page serves a page with a button to create a new room
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Poser</h1><form action=\"/\" method=\"POST\"><input type=\"submit\" value=\"New Room\"></form>"))
}

// NewRoomHandler just creates a UUID for a new room, then redirects the user.
//
// It would be cool to store all recent room IDs in a cookie,
// then render the list on the homepage for a user to return to.
// But for now we just forward them to a room URL, which will autogenerate the room.
func NewRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Homepage requested a new room
	//TODO could instead do some sort of TinyURL style Base58(SHA256(url, username))
	id := uuid.New()
	roomPath := fmt.Sprintf("/room/%s", id)
	http.Redirect(w, r, roomPath, http.StatusFound)
}

// RoomHandler serves the room assets
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(roomHTML))
}

// NameHandler allows a user to set their name via the API
func NameHandler() {
	// POST
	// Get name from JSON data
	// Add name,room to cookie
	// Reply with user data
}

func RoomUsersAPIHandler() {
	// GET
	// Get room ID from cookie
	// Get room users from DB
	// Reply with room data
}
