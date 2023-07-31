package main

import (
	"fmt"
	"html/template"
	"log"
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

// Could make this t := template.Must(template.ParseGlob("./templates/*.html"))
// then do t.ExecuteTemplate(w, "room.html", <data>)
// if more templates get added
var roomHTMLTemplate = template.Must(template.ParseFiles("./templates/room.html"))

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

// RoomHandler serves the room assets
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	data := struct{ Room string }{Room: vars["id"]}
	err := roomHTMLTemplate.Execute(w, data)
	if err != nil {
		log.Printf("failed to execute template: %s", err)
	}
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
