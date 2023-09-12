package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/dist/assets/"))))

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/", NewRoomHandler).Methods("POST")
	r.HandleFunc("/room/{id:.*}", RoomHandler)
	//r.HandleFunc("/gallery/{id:[0-9]+}", GalleryHandler)

	r.HandleFunc("/ws/{room}", HandleWebsocket)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
