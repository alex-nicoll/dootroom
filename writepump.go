package main

import (
	"github.com/gorilla/websocket"
)

// writePump runs a loop that copies a message from sendChan to the connection.
// It also listens for connection-specific errors and executes error handling.
func writePump(errSig *ErrorSignal, handleErr func(error), sendChan <-chan []byte, write Write) {
	for {
		select {
		case <-errSig.Done():
			handleErr(errSig.Err())
			return
		case message := <-sendChan:
			if err := write(websocket.BinaryMessage, message); err != nil {
				handleErr(err)
				return
			}
		}
	}
}
