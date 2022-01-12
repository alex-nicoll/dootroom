package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// ConnWriter exists to (mostly) decouple writePump from the websocket library.
type ConnWriter interface {
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// writePump pumps messages from send to conn.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer on a connection by
// executing all writes from this goroutine.
func writePump(send MsgChan, conn ConnWriter) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("writepump Close() " + err.Error())
		}
	}()
	for {
		message, ok := <-send
		if !ok {
			// The hub closed the channel.
			err := conn.WriteMessage(websocket.CloseMessage, []byte{})
			if err != nil {
				log.Println("writepump WriteMessage(CloseMessage) " + err.Error())
			}
			return
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			log.Println("writepump WriteMessage(BinaryMessage) " + err.Error())
			return
		}
	}
}
