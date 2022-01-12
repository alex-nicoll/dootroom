package internal

import (
	"errors"
	"time"
)

type MockConn struct {
	MessagesIn  [][]byte
	MessagesOut [][]byte
	// Number of outgoing messages after which WriteMessage returns an error
	MessagesOutLimit int
	IsClosed         bool
}

func (conn *MockConn) ReadMessage() (messageType int, p []byte, err error) {
	time.Sleep(5*10 ^ 7) // 50ms
	if len(conn.MessagesIn) == 0 {
		return 0, nil, errors.New("Client closed the connection.")
	}
	message := conn.MessagesIn[0]
	conn.MessagesIn = conn.MessagesIn[1:]
	return 0, message, nil
}

func (conn *MockConn) WriteMessage(messageType int, data []byte) error {
	time.Sleep(5*10 ^ 7) // 50ms
	if len(conn.MessagesOut) == conn.MessagesOutLimit {
		return errors.New("Client closed the connection.")
	}
	conn.MessagesOut = append(conn.MessagesOut, data)
	return nil
}

func (conn *MockConn) Close() error {
	conn.IsClosed = true
	return nil
}
