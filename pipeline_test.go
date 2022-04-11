package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func Test_pipeline(t *testing.T) {
	unmarshalOut := make(chan interface{})
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(unmarshalOut, modelChan)
	newConn := func() (in chan []byte, out chan []byte, re Read, wr Write, cl Close) {
		in = make(chan []byte)
		out = make(chan []byte)
		re = func() (messageType int, p []byte, err error) {
			return 0, <-in, nil
		}
		wr = func(messageType int, data []byte) error {
			out <- data
			return nil
		}
		cl = func() error {
			panic("Unexpected call to Close")
		}
		return
	}

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection.

	in1, out1, re1, wr1, cl1 := newConn()
	in2, out2, re2, wr2, cl2 := newConn()

	handleConn(re1, wr1, cl1)
	handleConn(re2, wr2, cl2)

	json := string(<-out1)
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-out2)
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When two Game of Life diffs (that is, the difference between the current
	// state and the desired state) come in as JSON on separate connections and
	// then a tick occurs, the pipeline should send the combined diff to each
	// connection.

	diff1 := "{\"30\":{\"30\":true},\"31\":{\"32\":true}}"
	diff2 := "{\"30\":{\"30\":true,\"31\":true},\"31\":{\"31\":true},\"32\":{\"31\":true}}"
	combinedDiff := "{\"30\":{\"30\":true,\"31\":true},\"31\":{\"31\":true,\"32\":true},\"32\":{\"31\":true}}"

	in1 <- []byte(diff1)
	in2 <- []byte(diff2)
	modelChan <- <-unmarshalOut
	modelChan <- <-unmarshalOut
	modelChan <- &Tick{}

	json = string(<-out1)
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-out2)
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a tick occurs, the pipeline should send the diff between the
	// current state and the previous state to each connection.

	diff := "{\"30\":{\"32\":true},\"31\":{\"31\":false},\"32\":{\"32\":true}}"

	modelChan <- &Tick{}

	json = string(<-out1)
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-out2)
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection. Unlike in the previous test for adding
	// new connections, there should now be non-empty cells in the GoL state.

	_, out3, re3, wr3, cl3 := newConn()

	handleConn(re3, wr3, cl3)

	json = string(<-out3)
	if !strings.HasPrefix(json, "[") || !strings.Contains(json, "true") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
}

// When invalid JSON comes in on a connection, a close message should be sent
// on the connection and then the connection should be closed.
func Test_invalidJSON(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff1(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("[]"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff2(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff3(t *testing.T) {
	invalidMessageTestTemplate(t, []byte(fmt.Sprintf("{\"0\":{\"%v\":true}}", GridDimY)))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff4(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{}}"))
}

// When reading from the connection returns an error that is not due to the
// client closing the connection, a close message should be sent and then the
// connection should be closed.
func Test_errorReadingMessage(t *testing.T) {
	handleConn := startPipeline()
	in := make(chan error)
	out := make(chan int)
	closed := make(chan struct{})
	handleConn(
		newReadErrorFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	<-out

	in <- errors.New("dummy error")

	verifyCloseMsgSentAndConnClosed(t, out, closed)
}

// When reading from the connection returns an error due to the client closing
// the connection, the connection should be closed and no close message should
// be sent.
func Test_errorReadingMessageClosedByClient(t *testing.T) {
	handleConn := startPipeline()
	in := make(chan error)
	out := make(chan struct{})
	closed := make(chan struct{})
	handleConn(
		newReadErrorFn(in, closed),
		func(messageType int, data []byte) error {
			out <- struct{}{}
			return nil
		},
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	<-out

	in <- &websocket.CloseError{}

	<-closed
	select {
	case <-out:
		t.Errorf("Unexpected message sent on connection.")
	default:
	}
}

// When writing to the connection returns an error, the connection should be
// closed.
func Test_errorWritingMessage(t *testing.T) {
	handleConn := startPipeline()
	closed := make(chan struct{})
	handleConn(
		func() (messageType int, p []byte, err error) {
			<-closed
			err = errors.New("Server closed the connection.")
			return
		},
		func(messageType int, data []byte) error {
			return errors.New("dummy error")
		},
		newCloseFn(closed),
	)

	// handleConn should have caused the GoL state initialization message to be
	// sent, thus producing an error.

	<-closed
}

// When the send buffer overflows, a close message should be sent on the
// connection and then the connection should be closed.
func Test_sendBufferOverflow(t *testing.T) {
	unmarshalOut := make(chan interface{})
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(unmarshalOut, modelChan)
	in := make(chan []byte)
	out := make(chan int)
	closed := make(chan struct{})
	errSig := handleConn(
		newReadPayloadFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)

	// This automaton runs forever, alternating between two states. This allows
	// us to produce an arbitrary number of outgoing messages.
	in <- []byte("{\"20\":{\"20\":true,\"21\":true,\"22\":true}}")
	modelChan <- <-unmarshalOut
	// Produce SendBufferLen+1 Ticks in order to overflow the send buffer.
	// writePump should currently be blocked trying to write the GoL state
	// initialization message to the connection, so no messages should be
	// pulled out of the send buffer.
	for i := 1; i <= SendBufferLen+1; i++ {
		modelChan <- &Tick{}
	}
	// Wait for the buffer to overflow.
	<-errSig.Done()

	// Unblock writePump. It may send some messages before sending the close
	// message because it randomly chooses whether to pull a message out of
	// the send buffer or read the error signal.
	for {
		if <-out == websocket.CloseMessage {
			break
		}
	}
	<-closed
}

// invalidMessageTestTemplate starts a pipeline with one connection, simulates
// the message coming in on the connection, then verifies that a close message
// was sent on the connection and the connection was closed.
func invalidMessageTestTemplate(t *testing.T, message []byte) {
	handleConn := startPipeline()
	in := make(chan []byte)
	out := make(chan int)
	closed := make(chan struct{})
	handleConn(
		newReadPayloadFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	<-out

	in <- message

	verifyCloseMsgSentAndConnClosed(t, out, closed)
}

func newReadPayloadFn(in chan []byte, closed chan struct{}) Read {
	return func() (messageType int, p []byte, err error) {
		select {
		case <-closed:
			err = errors.New("Server closed the connection.")
		case p = <-in:
		}
		return
	}
}

func newReadErrorFn(in chan error, closed chan struct{}) Read {
	return func() (messageType int, p []byte, err error) {
		err = <-in
		return
	}
}

func newWriteMessageTypeFn(out chan int) Write {
	return func(messageType int, data []byte) error {
		out <- messageType
		return nil
	}
}

func newCloseFn(closed chan struct{}) Close {
	return func() error {
		close(closed)
		return nil
	}
}

func verifyCloseMsgSentAndConnClosed(t *testing.T, out chan int, closed chan struct{}) {
	messageType := <-out
	if messageType != websocket.CloseMessage {
		t.Errorf("Expected CloseMessage but got message type %v", messageType)
	}
	<-closed
}

// TODO: test that error handling does not leak goroutines
// TODO: do these tests leak goroutines?
// TODO: add timeouts to help with debugging
