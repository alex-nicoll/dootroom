package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Read decouples the application from Conn.ReadMessage for testing purposes.
type Read = func() (messageType int, p []byte, err error)

// Write decouples the application from Conn.WriteMessage for testing purposes.
type Write = func(messageType int, data []byte) error

// Close decouples the application from Conn.Close for testing purposes.
type Close = func() error

const SendBufferLen = 256

// startPipeline runs clock, gol, and hub in separate goroutines and connects them in
// order via channels.  It returns a function that runs readPump in a goroutine
// that sends messages to gol, and runs writePump in a goroutine that receives
// messages from hub.
func startPipeline() func(Read, Write, Close) (*sync.WaitGroup, *ErrorSignal) {
	golChan := make(chan interface{})
	attachConn := startPipelineInternal(golChan, golChan)
	go clock(golChan)
	return attachConn
}

// Internal implementation of startPipeline exposed for testing purposes. This
// allows an additional stage to be added between readPump and gol, and omits
// clock so that tests can control gol via Tick messages.
func startPipelineInternal(readPumpOut chan interface{}, golChan chan interface{}) func(Read, Write, Close) (*sync.WaitGroup, *ErrorSignal) {
	hubChan := make(chan interface{})

	go gol(golChan, hubChan)
	go hub(hubChan)

	return func(re Read, wr Write, cl Close) (*sync.WaitGroup, *ErrorSignal) {
		// ErrorSignal for this connection
		errSig := NewErrorSignal()
		// Channel of messages to send on this connection
		sendChan := make(chan []byte, SendBufferLen)

		// Register this connection's send channel and ErrorSignal with the hub.
		listener := &Listener{sendChan, errSig}
		hubChan <- &Register{listener}

		// Tell gol to send down initialization data.
		golChan <- &InitListener{listener}

		handleErr := func(err error) {
			log.Println(err)
			if _, ok := err.(*BufferOverflowError); !ok {
				// If this was not a buffer overflow detected by the hub, then we need
				// to explicitly unregister.
				hubChan <- &Unregister{listener}
			}
			if !websocket.IsUnexpectedCloseError(err) {
				// The error is not due to the client closing the connection. Attempt
				// to send a close message. If this *was* a close error, then Gorilla
				// Websocket's default close handler should have already sent a close
				// message.
				err := wr(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Println(err)
				}
			}
			// Closing the connection should cause readPump to stop.
			cl()
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			writePump(errSig, handleErr, sendChan, wr)
		}()
		go func() {
			defer wg.Done()
			readPump(errSig, re, readPumpOut)
		}()
		return &wg, errSig
	}
}

// readPump runs a loop that reads a message from the connection, unmarshals
// JSON into a Diff, validates the Diff, and sends the Merge message to gol.
func readPump(errSig *ErrorSignal, read Read, golChan chan<- interface{}) {
	for {
		_, message, err := read()
		if err != nil {
			errSig.Close(err)
			return
		}
		diff := make(Diff)
		if err := json.Unmarshal(message, &diff); err != nil {
			errSig.Close(err)
			return
		}
		if err := validateDiff(diff); err != nil {
			errSig.Close(err)
			return
		}
		golChan <- &Merge{diff}
	}
}

func clock(golChan chan<- interface{}) {
	for {
		time.Sleep(170 * time.Millisecond)
		golChan <- &Tick{}
	}
}

type Merge struct {
	diff Diff
}

type InitListener struct {
	li *Listener
}

type Tick struct{}

// gol maintains the state of an instance of Conway's Game of Life, merging in
// changes from clients and propogating changes to hub to be broadcast to
// clients.
func gol(in <-chan interface{}, hubChan chan<- interface{}) {
	grid, diff := &Grid{}, make(Diff)

	// We could handle one Merge message and an arbitrary number of
	// InitListener messages concurrently. But for simplicity of implementation
	// we'll have one goroutine handle all three message types.
	for {
		switch m := (<-in).(type) {
		case *Merge:
			merge(m.diff, diff)
		case *InitListener:
			message, _ := json.Marshal(grid)
			hubChan <- &Forward{m.li, message}
		case *Tick:
			if len(diff) != 0 {
				message, _ := json.Marshal(diff)
				hubChan <- &Broadcast{message}
				flush(diff, grid)
			}
			nextState(grid, diff)
		}
	}
}

// Register a Listener
type Register struct {
	li *Listener
}

// Unregister a Listener
type Unregister struct {
	li *Listener
}

// Broadcast a websocket message to all registered Listeners
type Broadcast struct {
	message []byte
}

// Forward a websocket message to a Specific listener
type Forward struct {
	li      *Listener
	message []byte
}

// hub runs a loop that sends websocket messages to Listeners.
func hub(in <-chan interface{}) {
	listeners := make(map[*Listener]bool)
	for {
		switch m := (<-in).(type) {
		case *Register:
			listeners[m.li] = true
		case *Unregister:
			delete(listeners, m.li)
		case *Broadcast:
			for li := range listeners {
				select {
				case li.sendChan <- m.message:
				default:
					li.errSig.Close(&BufferOverflowError{})
					delete(listeners, li)
				}
			}
		case *Forward:
			select {
			case m.li.sendChan <- m.message:
			default:
				m.li.errSig.Close(&BufferOverflowError{})
				delete(listeners, m.li)
			}
		}
	}
}

// writePump runs a loop that copies a message from sendChan to the connection,
// or executes error handling when a connection-specific error is detected.
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
