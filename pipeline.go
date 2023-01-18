package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// readFromConn decouples the application from websocket.Conn.ReadMessage for testing purposes.
type readFromConn = func() (messageType int, p []byte, err error)

// writeToConn decouples the application from websocket.Conn.WriteMessage for testing purposes.
type writeToConn = func(messageType int, data []byte) error

// closeConn decouples the application from websocket.Conn.closeConn for testing purposes.
type closeConn = func() error

const sendBufferLen = 256

type pipeline struct {
	readPumpOut chan interface{}
	golChan     chan interface{}
	hubChan     chan interface{}
}

// startPipeline runs clock, gol, and hub in separate goroutines and connects
// them in that order via channels.
func startPipeline() *pipeline {
	golChan := make(chan interface{})
	pl := startPipelineInternal(golChan, golChan)
	go clock(golChan)
	return pl
}

// Internal implementation of startPipeline exposed for testing purposes. It
// allows an additional stage to be added between readPump and gol, and omits
// clock so that tests can control gol via tick messages.
func startPipelineInternal(readPumpOut chan interface{}, golChan chan interface{}) *pipeline {
	hubChan := make(chan interface{})
	go gol(golChan, hubChan)
	go hub(hubChan)
	return &pipeline{readPumpOut, golChan, hubChan}

}

// attachConn attaches a connection to a pipeline. It starts readPump in a
// goroutine that sends messages to gol, and starts writePump in a goroutine
// that receives messages from hub. It also causes initialization data to be
// sent to the client. For testing purposes, attachConn returns the errorSignal
// associated with the connection and a WaitGroup that can be used to wait for
// writePump and readPump to stop.
func attachConn(pl *pipeline, re readFromConn, wr writeToConn, cl closeConn) (*sync.WaitGroup, *errorSignal) {
	// errorSignal for this connection
	errSig := newErrorSignal()
	// Channel of messages to send on this connection
	sendChan := make(chan []byte, sendBufferLen)

	// Register this connection's send channel and errorSignal with the hub.
	li := &listener{sendChan, errSig}
	pl.hubChan <- &register{li}

	// Tell gol to send down initialization data.
	pl.golChan <- &initListener{li}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errHan := &errorHandler{pl.hubChan, li, wr, cl}
		writePump(errSig, errHan, sendChan, wr)
	}()
	go func() {
		defer wg.Done()
		readPump(errSig, re, pl.readPumpOut)
	}()
	return &wg, errSig
}

type errorHandler struct {
	hubChan chan interface{}
	li      *listener
	wr      writeToConn
	cl      closeConn
}

func (errHan *errorHandler) run(err error) {
	log.Println(err)
	if _, ok := err.(*bufferOverflowError); !ok {
		// If this was not a buffer overflow detected by the hub, then we need
		// to explicitly unregister.
		errHan.hubChan <- &unregister{errHan.li}
	}
	if !websocket.IsUnexpectedCloseError(err) {
		// The error is not due to the client closing the connection. Attempt
		// to send a close message. If this *was* a close error, then Gorilla
		// Websocket's default close handler should have already sent a close
		// message.
		err := errHan.wr(websocket.CloseMessage, []byte{})
		if err != nil {
			log.Printf("Error sending close message: %v\n", err)
		}
	}
	// Close the connection. This should cause readPump to stop if it
	// hasn't already.
	if err := errHan.cl(); err != nil {
		log.Printf("Error closing connection: %v\n", err)
	}
}

// readPump runs a loop that reads a message from the connection, unmarshals
// JSON into a diff, validates the diff, and sends the mergeDiff message to
// gol.
func readPump(errSig *errorSignal, read readFromConn, golChan chan<- interface{}) {
	for {
		_, message, err := read()
		if err != nil {
			errSig.send(err)
			return
		}
		df := make(diff)
		if err := json.Unmarshal(message, &df); err != nil {
			errSig.send(err)
			return
		}
		if err := validateDiff(df); err != nil {
			errSig.send(err)
			return
		}
		golChan <- &mergeDiff{df}
	}
}

func clock(golChan chan<- interface{}) {
	for {
		time.Sleep(170 * time.Millisecond)
		golChan <- &tick{}
	}
}

type mergeDiff struct {
	df diff
}

type initListener struct {
	li *listener
}

type tick struct{}

// gol maintains the state of an instance of Conway's Game of Life, merging in
// changes from clients and propogating changes to hub to be broadcast to
// clients. See protocol.md for more context regarding the implementation.
func gol(in <-chan interface{}, hubChan chan<- interface{}) {
	g, df := &grid{}, make(diff)

	// isEmptyDiffSent is true if the grid has stopped evolving (because it is
	// empty or consists entirely of still lifes), and we have broadcasted a
	// single empty diff to indicate that the stream of messages has ended.
	isEmptyDiffSent := false

	// We could handle one mergeDiff message and an arbitrary number of
	// initListener messages concurrently. But for simplicity of implementation
	// we'll have one goroutine handle all three message types.
	for {
		switch m := (<-in).(type) {
		case *mergeDiff:
			merge(m.df, df)
		case *initListener:
			gridMessage, _ := json.Marshal(g)
			hubChan <- &forward{m.li, gridMessage}
			if isEmptyDiffSent {
				// Send the empty diff to this new Listener as well.
				emptyDiffMessage, _ := json.Marshal(df)
				hubChan <- &forward{m.li, emptyDiffMessage}
			}
		case *tick:
			if len(df) != 0 {
				message, _ := json.Marshal(df)
				hubChan <- &broadcast{message}
				flush(df, g)
				nextState(g, df)
				isEmptyDiffSent = false
			} else if !isEmptyDiffSent {
				message, _ := json.Marshal(df)
				hubChan <- &broadcast{message}
				isEmptyDiffSent = true
			}
			// Note: Using len(diff) to determine whether the grid has stopped
			// evolving hinges on the assumption that it isn't possible to
			// "cancel out" a diff via a mergeDiff message. So care would need
			// to be taken if we were to implement, e.g., a way for the client
			// to erase grid cells.
		}
	}
}

// register a Listener
type register struct {
	li *listener
}

// unregister a Listener
type unregister struct {
	li *listener
}

// broadcast a websocket message to all registered Listeners
type broadcast struct {
	message []byte
}

// forward a websocket message to a specific Listener
type forward struct {
	li      *listener
	message []byte
}

// hub runs a loop that sends websocket messages to Listeners.
func hub(in <-chan interface{}) {
	listeners := make(map[*listener]bool)
	for {
		switch m := (<-in).(type) {
		case *register:
			listeners[m.li] = true
		case *unregister:
			delete(listeners, m.li)
		case *broadcast:
			for li := range listeners {
				select {
				case li.sendChan <- m.message:
				default:
					li.errSig.send(&bufferOverflowError{})
					delete(listeners, li)
				}
			}
		case *forward:
			select {
			case m.li.sendChan <- m.message:
			default:
				m.li.errSig.send(&bufferOverflowError{})
				delete(listeners, m.li)
			}
		}
	}
}

// writePump runs a loop that copies a message from sendChan to the connection,
// or executes error handling when a connection-specific error is detected.
func writePump(errSig *errorSignal, errHan *errorHandler, sendChan <-chan []byte, write writeToConn) {
	for {
		select {
		case <-errSig.signal():
			errHan.run(errSig.err())
			return
		case message := <-sendChan:
			if err := write(websocket.BinaryMessage, message); err != nil {
				errHan.run(err)
				return
			}
		}
	}
}
