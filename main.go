package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Serve /static/ from ./static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/dist/assets/"))))

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/", NewRoomHandler).Methods("POST")
	//r.HandleFunc("/room/{id:[0-9]+}", RoomHandler)
	r.HandleFunc("/room/{id:.*}", RoomHandler)
	//r.HandleFunc("/gallery/{id:[0-9]+}", GalleryHandler)

	server := NewServer()
	r.HandleFunc("/ws/{room}", server.Echo)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
