package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func Test_pipeline(t *testing.T) {
	readPumpOut := make(chan interface{})
	golChan := make(chan interface{})
	attachConn := startPipelineInternal(readPumpOut, golChan)

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection.

	in1, out1, re1, wr1, cl1 := newConn(t)
	in2, out2, re2, wr2, cl2 := newConn(t)

	attachConn(re1, wr1, cl1)
	attachConn(re2, wr2, cl2)

	json := string(recv(t, out1))
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(recv(t, out2))
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When two Game of Life diffs (that is, the difference between the current
	// state and the desired state) come in as JSON on separate connections and
	// then a tick occurs, the pipeline should send the combined diff to each
	// connection.

	diff1 := "{\"30\":{\"30\":\"#aaaaaa\"},\"31\":{\"32\":\"#aaaaaa\"}}"
	diff2 := "{\"30\":{\"30\":\"#aaaaaa\",\"31\":\"#aaaaaa\"},\"31\":{\"31\":\"#aaaaaa\"},\"32\":{\"31\":\"#aaaaaa\"}}"
	combinedDiff := "{\"30\":{\"30\":\"#aaaaaa\",\"31\":\"#aaaaaa\"},\"31\":{\"31\":\"#aaaaaa\",\"32\":\"#aaaaaa\"},\"32\":{\"31\":\"#aaaaaa\"}}"

	send(t, in1, []byte(diff1))
	send(t, in2, []byte(diff2))
	forward(t, golChan, readPumpOut)
	forward(t, golChan, readPumpOut)
	send[interface{}](t, golChan, &Tick{})

	json = string(recv(t, out1))
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(recv(t, out2))
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a tick occurs, the pipeline should send the diff between the
	// current state and the previous state to each connection.

	diff := "{\"30\":{\"32\":\"#aaaaaa\"},\"31\":{\"31\":\"\"},\"32\":{\"32\":\"#aaaaaa\"}}"

	send[interface{}](t, golChan, &Tick{})

	json = string(recv(t, out1))
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(recv(t, out2))
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection. Unlike in the previous test for adding
	// new connections, there should now be non-empty cells in the GoL state.

	_, out3, re3, wr3, cl3 := newConn(t)

	attachConn(re3, wr3, cl3)

	json = string(recv(t, out3))
	if !strings.HasPrefix(json, "[") || !strings.Contains(json, "\"#aaaaaa\"") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
}

func Test_pipelineEmptyDiff(t *testing.T) {
	readPumpOut := make(chan interface{})
	golChan := make(chan interface{})
	attachConn := startPipelineInternal(readPumpOut, golChan)

	in1, out1, re1, wr1, cl1 := newConn(t)
	_, out2, re2, wr2, cl2 := newConn(t)

	attachConn(re1, wr1, cl1)
	attachConn(re2, wr2, cl2)

	// Handle the GoL state initialization message
	recv(t, out1)
	recv(t, out2)

	// When the Game of Life state stops changing, the pipeline should send
	// an empty diff to each connection. We can check this by sending a diff
	// containing a single live cell to the pipeline. For each connection, we
	// then expect to see a single-cell life diff, single-cell death diff, and
	// finally an empty diff.

	diff := "{\"0\":{\"0\":\"#aaaaaa\"}}"

	send(t, in1, []byte(diff))
	forward(t, golChan, readPumpOut)
	send[interface{}](t, golChan, &Tick{})
	recv(t, out1)
	recv(t, out2)
	send[interface{}](t, golChan, &Tick{})
	recv(t, out1)
	recv(t, out2)
	send[interface{}](t, golChan, &Tick{})

	json := string(recv(t, out1))
	if json != "{}" {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(recv(t, out2))
	if json != "{}" {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// After the pipeline sends an empty diff to each connection, it should
	// send no further diffs until the state changes again. We can check this
	// by sending the single-cell life diff again and verifying that it is the next
	// diff received on each connection.

	send(t, in1, []byte(diff))
	forward(t, golChan, readPumpOut)
	send[interface{}](t, golChan, &Tick{})

	json = string(recv(t, out1))
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(recv(t, out2))
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// Clean up; receive the single-cell death diff and the empty diff.
	send[interface{}](t, golChan, &Tick{})
	recv(t, out1)
	recv(t, out2)
	send[interface{}](t, golChan, &Tick{})
	recv(t, out1)
	recv(t, out2)

	// After the the pipeline sends an empty diff to each connection, when a
	// new connection is made, the pipeline should send the empty diff to that
	// connection.

	_, out3, re3, wr3, cl3 := newConn(t)
	attachConn(re3, wr3, cl3)
	// Handle the GoL state initialization message
	recv(t, out3)

	json = string(recv(t, out3))
	if json != "{}" {
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
	invalidMessageTestTemplate(t, []byte(fmt.Sprintf("{\"0\":{\"%v\":\"#aaaaaa\"}}", GridDimY)))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff4(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff5(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"#123\"}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff6(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"#123xyz\"}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff7(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"a#123abc\"}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff8(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"#123abcc\"}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff9(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"#123abc#123abc\"}}"))
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff10(t *testing.T) {
	invalidMessageTestTemplate(t, []byte("{\"0\":{\"0\":\"\"}}"))
}

// When reading from the connection returns an error that is not due to the
// client closing the connection, a close message should be sent and then the
// connection should be closed.
func Test_errorReadingMessage(t *testing.T) {
	attachConn := startPipeline()
	in := make(chan error)
	out := make(chan int)
	closed := make(chan struct{})
	attachConn(
		newReadErrorFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	recv(t, out)

	send(t, in, errors.New("dummy error"))

	verifyCloseMsgSentAndConnClosed(t, out, closed)
}

// When reading from the connection returns an error due to the client closing
// the connection, the connection should be closed and no close message should
// be sent.
func Test_errorReadingMessageClosedByClient(t *testing.T) {
	attachConn := startPipeline()
	in := make(chan error)
	out := make(chan struct{})
	closed := make(chan struct{})
	attachConn(
		newReadErrorFn(in, closed),
		func(messageType int, data []byte) error {
			out <- struct{}{}
			return nil
		},
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	recv(t, out)

	send[error](t, in, &websocket.CloseError{})

	recv(t, closed)
	select {
	case <-out:
		t.Errorf("Unexpected message sent on connection.")
	default:
	}
}

// When writing to the connection returns an error, the connection should be
// closed.
func Test_errorWritingMessage(t *testing.T) {
	attachConn := startPipeline()
	closed := make(chan struct{})
	attachConn(
		newReadUntilClosedFn(closed),
		func(messageType int, data []byte) error {
			return errors.New("dummy error")
		},
		newCloseFn(closed),
	)

	// attachConn should have caused the GoL state initialization message to be
	// sent, thus producing an error.

	recv(t, closed)
}

// When the send buffer overflows, a close message should be sent on the
// connection and then the connection should be closed.
func Test_sendBufferOverflow(t *testing.T) {
	readPumpOut := make(chan interface{})
	modelChan := make(chan interface{})
	attachConn := startPipelineInternal(readPumpOut, modelChan)
	in := make(chan []byte)
	out := make(chan int)
	closed := make(chan struct{})
	_, errSig := attachConn(
		newReadPayloadFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)

	// This automaton runs forever, alternating between two states. This allows
	// us to produce an arbitrary number of outgoing messages.
	send(t, in, []byte("{\"20\":{\"20\":\"#aaaaaa\",\"21\":\"#aaaaaa\",\"22\":\"#aaaaaa\"}}"))
	forward(t, modelChan, readPumpOut)
	// Produce SendBufferLen+1 Ticks in order to overflow the send buffer.
	// writePump should currently be blocked trying to write the GoL state
	// initialization message to the connection, so no messages should be
	// pulled out of the send buffer.
	for i := 1; i <= SendBufferLen+1; i++ {
		send[interface{}](t, modelChan, &Tick{})
	}
	// Wait for the buffer to overflow.
	recv(t, errSig.Done())

	// Unblock writePump. It may send some messages before sending the close
	// message because it randomly chooses whether to pull a message out of
	// the send buffer or read the error signal.
	for {
		m := recv(t, out)
		if m == websocket.CloseMessage {
			break
		}
	}
	recv(t, closed)
}

// When an error occurs related to a connection and resources are cleaned up,
// no goroutines should be leaked.
func Test_leak(t *testing.T) {
	attachConn := startPipeline()
	closed := make(chan struct{})
	wg, errSig := attachConn(
		newReadUntilClosedFn(closed),
		func(messageType int, data []byte) error {
			return nil
		},
		newCloseFn(closed),
	)

	errSig.Close(errors.New("dummy error"))

	wg.Wait()
}

func newConn(t *testing.T) (in chan []byte, out chan []byte, re Read, wr Write, cl Close) {
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
		t.Errorf("Unexpected call to Close")
		return nil
	}
	return
}

// invalidMessageTestTemplate starts a pipeline with one connection, simulates
// the message coming in on the connection, then verifies that a close message
// was sent on the connection and the connection was closed.
func invalidMessageTestTemplate(t *testing.T, message []byte) {
	attachConn := startPipeline()
	in := make(chan []byte)
	out := make(chan int)
	closed := make(chan struct{})
	attachConn(
		newReadPayloadFn(in, closed),
		newWriteMessageTypeFn(out),
		newCloseFn(closed),
	)
	// Handle the GoL state initialization message
	recv(t, out)

	send(t, in, message)

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

func newReadUntilClosedFn(closed chan struct{}) Read {
	return func() (messageType int, p []byte, err error) {
		<-closed
		err = errors.New("Server closed the connection.")
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
	messageType := recv(t, out)
	if messageType != websocket.CloseMessage {
		t.Errorf("Expected CloseMessage but got message type %v", messageType)
	}
	recv(t, closed)
}

// recv performs the channel receive operation with a timeout.
func recv[U any](t *testing.T, ch <-chan U) U {
	select {
	case <-time.After(2 * time.Second):
		t.Errorf("Channel receive operation timed out")
		var zero U
		return zero
	case v := <-ch:
		return v
	}
}

// send performs the channel send operation with a timeout.
func send[U any](t *testing.T, ch chan<- U, v U) {
	select {
	case <-time.After(2 * time.Second):
		t.Errorf("Channel send operation timed out")
	case ch <- v:
	}
	return
}

// forward copies a value from chOut to chIn, with a timeout for each channel
// operation.
func forward[U any](t *testing.T, chIn chan<- U, chOut <-chan U) {
	send(t, chIn, recv(t, chOut))
}
