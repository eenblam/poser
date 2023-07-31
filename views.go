package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// The home page serves a page with a button to create a new room
//
// TODO it should also include a form to instead join an existing room by ID
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Poser</h1><form action=\"/\" method=\"POST\"><input type=\"submit\" value=\"New Room\"></form>"))
}

func NewRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Homepage requested a new room
	//TODO could instead do some sort of TinyURL style Base58(SHA256(url, username))
	id := uuid.New()
	// Create a room
	// Redirect to room page
	//w.Write([]byte("Gonna make you a room!\n"))
	roomPath := fmt.Sprintf("/room/%s", id)
	http.Redirect(w, r, roomPath, http.StatusFound)
}

var roomHTMLTemplate = `<title>Room: %v</title>
<body>
<p>Room ID: %v</p>
<script src="/static/websocket.js"></script>
</body> `

//<script type="text/javascript" src="/static/websocket.js"></script>
//<script type="module" src="https://unpkg.com/simple-peer@9.11.1/simplepeer.min.js"></script>

// RoomHandler serves the room assets
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, roomHTMLTemplate, vars["id"], vars["id"])
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
