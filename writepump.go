package main

import (
	"github.com/gorilla/websocket"
)

// writePump runs a loop that copies messages from sendChan to the connection.
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
