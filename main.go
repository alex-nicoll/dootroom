package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// MsgChan is a channel of WebSocket messages
type MsgChan = chan []byte

// TODO: Pick an appropriate buffer size. See Gorilla WebSocket documentation.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			serveWebSocket(hub, w, r)
			return
		}
		http.ServeFile(w, r, "./main.html")
	})
	http.HandleFunc("/main.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./main.js")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
	// TODO: Pick an appropriate port number.
}

func serveWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	send := make(MsgChan, 256)
	hub.register <- send
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go writePump(send, conn)
	go readPump(conn, hub.broadcast, func() { hub.unregister <- send })
	// TODO: Handle control messages. See Gorilla WebSocket documentation.
}
