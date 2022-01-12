package main

import (
	"log"
)

// ConnReader exists to decouple readPump from the websocket library.
type ConnReader interface {
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// readPump pumps messages from conn to broadcast.
//
// A goroutine running readPump is started for each connection. The
// application ensures that there is at most one reader on a connection by
// executing all reads from this goroutine.
func readPump(conn ConnReader, broadcast MsgChan, unregister func()) {
	defer func() {
		unregister()
		err := conn.Close()
		if err != nil {
			log.Println("readpump Close() " + err.Error())
		}
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("readpump ReadMessage() " + err.Error())
			return
		}
		broadcast <- message
	}
}
