package main

import (
	"log"

	"github.com/gorilla/websocket"
)

const SendBufferLen = 256

// startPipeline runs model and hub in separate goroutines and connects them in
// order via channels to form a pipeline. It also runs clock in a goroutine and
// connects it to model.
//
// It returns a function that connects a WebSocket connection to the pipeline.
// The returned function runs readPump and writePump in new goroutines.
func startPipeline() func(Read, Write, Close) *ErrorSignal {
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(modelChan, modelChan)
	go clock(modelChan)
	return handleConn
}

// Internal implementation of startPipeline exposed for testing purposes. This
// allows an additional stage to be added between readPump and model, and
// omits clock so that tests can control model via Tick messages.
func startPipelineInternal(readPumpOut chan interface{}, modelChan chan interface{}) func(Read, Write, Close) *ErrorSignal {
	hubChan := make(chan interface{})

	go model(modelChan, hubChan)
	go hub(hubChan)

	return func(re Read, wr Write, cl Close) (errSig *ErrorSignal) {
		// ErrorSignal for this connection
		errSig = NewErrorSignal()
		// Channel of messages to send on this connection
		sendChan := make(chan []byte, SendBufferLen)

		// Register this connection's send channel and ErrorSignal with the hub.
		listener := &Listener{sendChan, errSig}
		hubChan <- &Register{listener}

		// Tell the model to send down initialization data.
		modelChan <- &InitListener{listener}

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

		go writePump(errSig, handleErr, sendChan, wr)
		go readPump(errSig, re, readPumpOut)
		return
	}
}
