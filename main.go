package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/", Handle)
	m.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/dist/assets/"))))
	log.Fatal(http.ListenAndServe(":8080", m))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	switch p := r.URL.Path; {
	case p == "/": // home
		HomeHandler(w, r)
		return
	case strings.HasPrefix(p, "/room/"):
		RoomHandler(w, r)
		return
	case strings.HasPrefix(p, "/ws/"):
		HandleWebsocket(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
