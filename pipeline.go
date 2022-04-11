package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

const SendBufferLen = 256

// Conn decouples startPipeline from websocket.Conn for testing purposes.
type Conn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// startPipeline runs unmarshal, model, and hub in separate goroutines and
// connects them in order via channels to form a pipeline. It also runs clock
// in a goroutine and connects it to model.
//
// It returns a function that connects a WebSocket connection to the pipeline.
// The returned function runs readPump and writePump in new goroutines, and
// handles errors related to the connection in a third goroutine.
func startPipeline() func(Conn) *ErrorSignal {
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(modelChan, modelChan)
	go clock(modelChan)
	return handleConn
}

// Internal implementation of startPipeline exposed for testing purposes. This
// allows an additional stage to be added between unmarshal and model, and
// omits clock so that tests can control model via Tick messages.
func startPipelineInternal(unmarshalOut chan interface{}, modelChan chan interface{}) func(Conn) *ErrorSignal {
	unmarshalChan := make(chan interface{})
	hubChan := make(chan interface{})

	go unmarshal(unmarshalChan, unmarshalOut)
	go model(modelChan, hubChan)
	go hub(hubChan)

	return func(conn Conn) (errSig *ErrorSignal) {
		// ErrorSignal for this connection
		errSig = NewErrorSignal()
		go func() {
			// TODO: Handle control messages. See Gorilla WebSocket documentation.
			// TODO: Once this is tested, see if refactoring to pass error handling as a function to writePump makes things simpler

			// Channel of messages to send on this connection
			sendChan := make(chan []byte, SendBufferLen)

			// Register this connection's send channel and ErrorSignal with the hub.
			listener := &Listener{sendChan, errSig}
			hubChan <- &Register{listener}

			// Tell the model to send down initialization data.
			modelChan <- &InitListener{listener}

			// Start a goroutine to copy messages from the send channel to the
			// connection.
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				writePump(errSig, sendChan, conn)
			}()
			// Start a goroutine to copy messages from the connection to the unmarshal
			// channel.
			go readPump(errSig, conn, unmarshalChan)

			// Wait for an error.
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

			// Closing the connection should cause readPump to stop if it is
			// currently blocked waiting to read a message from the connection.
			conn.Close()
		}()
		return
	}
}
