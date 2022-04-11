package main

import (
	"github.com/gorilla/websocket"
)

// ConnWriter decouples writePump from websocket.Conn for testing purposes.
type ConnWriter interface {
	WriteMessage(messageType int, data []byte) error
}

// writePump runs a loop that copies messages from sendChan to conn.
func writePump(errSig *ErrorSignal, handleErr func(error), sendChan <-chan []byte, conn ConnWriter) {
	for {
		select {
		case <-errSig.Done():
			handleErr(errSig.Err())
			return
		case message := <-sendChan:
			if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				errSig.Close(err)
				handleErr(errSig.Err())
				return
			}
		}
	}
}
