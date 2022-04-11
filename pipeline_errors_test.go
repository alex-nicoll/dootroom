package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gorilla/websocket"
)

type mockConnThatCloses struct {
	in     chan interface{}
	out    chan *outMessage
	closed chan struct{}
}

type outMessage struct {
	messageType int
	data        []byte
}

func (conn *mockConnThatCloses) ReadMessage() (messageType int, p []byte, err error) {
	select {
	case <-conn.closed:
		return 0, nil, errors.New("Server closed the connection.")
	case m := <-conn.in:
		switch m := m.(type) {
		case []byte:
			return 0, m, nil
		case *websocket.CloseError:
			return 0, nil, m
		case error:
			return 0, nil, m
		default:
			panic("Unexpected type sent on channel conn.in")
		}
	}
}

func (conn *mockConnThatCloses) WriteMessage(messageType int, data []byte) error {
	conn.out <- &outMessage{messageType, data}
	return nil

}

func (conn *mockConnThatCloses) Close() error {
	close(conn.closed)
	return nil
}

func startPipelineWithOneConn() *mockConnThatCloses {
	handleConn := startPipeline()
	conn := &mockConnThatCloses{
		in:     make(chan interface{}),
		out:    make(chan *outMessage),
		closed: make(chan struct{}),
	}
	handleConn(conn)
	// Handle the GoL state initialization message
	<-conn.out
	return conn
}

func verifyCloseMsgSentAndConnClosed(t *testing.T, conn *mockConnThatCloses) {
	m := <-conn.out
	if m.messageType != websocket.CloseMessage {
		t.Errorf("Expected CloseMessage but got message type %v", m.messageType)
	}
	<-conn.closed
}

// When invalid JSON comes in on a connection, a close message should be sent
// on the connection and then the connection should be closed.
func Test_invalidJSON(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- []byte("{")

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff1(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- []byte("[]")

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff2(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- []byte("{}")

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff3(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- []byte(fmt.Sprintf("{\"0\":{\"%v\":true}}", GridDimY))

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When valid JSON that is not a valid Game of Life diff comes in on a
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_invalidDiff4(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- []byte("{\"0\":{}}")

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When ReadMessage returns an error that is not due to the client closing the
// connection, a close message should be sent on the connection and then the
// connection should be closed.
func Test_errorReadingMessage(t *testing.T) {
	conn := startPipelineWithOneConn()

	conn.in <- errors.New("dummy error")

	verifyCloseMsgSentAndConnClosed(t, conn)
}

// When ReadMessage returns an error due to the client closing the connection,
// the connection should be closed and no close message should be sent.
func Test_closedByClient(t *testing.T) {
	handleConn := startPipeline()
	conn := &mockConnThatCloses{
		in:     make(chan interface{}),
		out:    make(chan *outMessage, 1),
		closed: make(chan struct{}),
	}
	handleConn(conn)
	// Handle the GoL state initialization message
	<-conn.out

	conn.in <- &websocket.CloseError{}

	<-conn.closed
	select {
	case m := <-conn.out:
		t.Errorf("Unexpected message sent on connection: %v", m)
	default:
	}
}

// When the send buffer overflows, a close message should be sent on the
// connection and then the connection should be closed.
func Test_sendBufferOverflow(t *testing.T) {
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(modelChan, modelChan)
	conn := &mockConnThatCloses{
		in:     make(chan interface{}),
		out:    make(chan *outMessage),
		closed: make(chan struct{}),
	}
	errSig := handleConn(conn)

	// This automaton runs forever, alternating between two states. This allows
	// us to produce an arbitrary number of outgoing messages.
	json := []byte("{\"20\":{\"20\":true,\"21\":true,\"22\":true}}")
	conn.in <- json
	// To overflow the buffer, we need SendBufferLen+2 Ticks as opposed to +1
	// because one message should be pulled out of the buffer by writePump,
	// which should then block trying to write to the connection.
	for i := 1; i < SendBufferLen+2; i++ {
		modelChan <- &Tick{}
	}
	// Wait for the buffer to overflow
	<-errSig.Done()

	// Unblock writePump. It may send some messages before sending the close
	// message because it randomly chooses whether to pull a message out of
	// the send buffer or read the error signal.
	for {
		m := <-conn.out
		if m.messageType == websocket.CloseMessage {
			break
		}
	}
	<-conn.closed
}

// TODO: test that error handling does not leak goroutines
// TODO: do these tests leak goroutines?
// TODO: use functions instead of doing this method overriding thing so that my test functions can go in any file. OR have a mockConn where the method implementations just call functions stored in the mockConn. Then the specific test can define the implementation.
// TODO: add timeouts to help with debugging
