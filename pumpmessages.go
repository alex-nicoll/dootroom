package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Conn decouples pumpMessages from websocket.Conn for testing purposes.
type Conn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// pumpMessages starts a readPump goroutine and a writePump goroutine for the
// WebSocket connection and sets up communication with the hub goroutine. It
// then waits for an error to occur related to the connection and performs
// error handling in the current goroutine.
func pumpMessages(hubChan chan interface{}, conn Conn) {
	// TODO: Handle control messages. See Gorilla WebSocket documentation.

	// ErrorSignal for this connection
	errSig := NewErrorSignal()
	// Channel of messages to send on this connection
	sendChan := make(chan []byte, 256)

	// Register this connection's send channel and ErrorSignal with the hub.
	listener := &Listener{sendChan, errSig}
	hubChan <- &Register{listener}

	var wg sync.WaitGroup
	// Start a goroutine to copy messages from the send channel to the
	// connection.
	wg.Add(1)
	go func() {
		defer wg.Done()
		writePump(errSig, sendChan, conn)
	}()
	// Start a goroutine to copy messages from the connection to the hub
	// channel.
	wg.Add(1)
	go func() {
		defer wg.Done()
		readPump(errSig, conn, hubChan)
	}()

	// Wait for the writePump, readPump, or hub goroutine to detect an error.
	<-errSig.Done()
	err := errSig.Err()
	log.Println(err)

	if _, ok := err.(*BufferOverflowError); !ok {
		// If this was not a buffer overflow detected by the hub, then we need
		// to explicitly unregister.
		hubChan <- &Unregister{listener}
	}

	wg.Wait()

	if !websocket.IsUnexpectedCloseError(err) {
		// The error is not due to the client closing the connection. Attempt
		// to send a close message. If this *was* a close error, then Gorilla
		// Websocket's default close handler should have already sent a close
		// message.
		err := conn.WriteMessage(websocket.CloseMessage, []byte{})
		if err != nil {
			log.Println(err)
		}
	}

	conn.Close()
}
