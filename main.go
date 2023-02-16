// Command multi-life runs a multiplayer Conway's Game of Life server.
// See README.md for installation and usage instructions.
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	pl := startPipeline()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !websocket.IsWebSocketUpgrade(r) {
			serveFileNoCache(w, r, "./assets/main.html")
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		attachConn(
			pl,
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
		serveFileNoCache(w, r, "./assets/main.js")
	})
	http.HandleFunc("/main.css", func(w http.ResponseWriter, r *http.Request) {
		serveFileNoCache(w, r, "./assets/main.css")
	})
	http.HandleFunc("/beehive_oscillator.png", func(w http.ResponseWriter, r *http.Request) {
		serveFileNoCache(w, r, "./assets/beehive_oscillator.png")
	})
	log.Fatal(http.ListenAndServe(":80", nil))
}

// serveFileNoCache serves a file and directs the client to always request the
// most up-to-date version.
func serveFileNoCache(w http.ResponseWriter, r *http.Request, name string) {
	// "no-store" prevents clients from storing the response. A more efficient
	// but complicated approach would be to allow clients to store the
	// response, but have them check that it's up-to-date before using it. See
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#force_revalidation
	w.Header()["Cache-Control"] = []string{"no-store"}
	http.ServeFile(w, r, name)
}
