package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	attachConn := startPipeline()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !websocket.IsWebSocketUpgrade(r) {
			http.ServeFile(w, r, "./main.html")
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		attachConn(
			func() (messageType int, p []byte, err error) {
				return conn.ReadMessage()
			},
			func(messageType int, data []byte) error {
				return conn.WriteMessage(messageType, data)
			},
			func() error {
				return conn.Close()
			},
		)
	})
	http.HandleFunc("/main.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./main.js")
	})
	http.HandleFunc("/main.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./main.css")
	})
	http.HandleFunc("/info_icon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./info_icon.svg")
	})
	http.HandleFunc("/close_icon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./close_icon.svg")
	})
	http.HandleFunc("/beehive_oscillator.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./beehive_oscillator.png")
	})
	log.Fatal(http.ListenAndServe(":"+os.Args[1], nil))
}
