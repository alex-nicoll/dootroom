package internal

import (
	"time"
)

type MockCloseError struct{}

func (err *MockCloseError) Error() string {
	return "Client closed the connection."
}

type MockConn struct {
	MessagesIn  [][]byte
	MessagesOut [][]byte
	// Number of outgoing messages after which WriteMessage returns an error
	MessagesOutLimit int
	IsClosed         bool
	OnReadMessage    func()
	OnWriteMessage   func()
}

func (conn *MockConn) ReadMessage() (messageType int, p []byte, err error) {
	if conn.OnReadMessage != nil {
		conn.OnReadMessage()
	}
	time.Sleep(50 * time.Millisecond)
	if len(conn.MessagesIn) == 0 {
		return 0, nil, &MockCloseError{}
	}
	message := conn.MessagesIn[0]
	conn.MessagesIn = conn.MessagesIn[1:]
	return 0, message, nil
}

func (conn *MockConn) WriteMessage(messageType int, data []byte) error {
	if conn.OnWriteMessage != nil {
		conn.OnWriteMessage()
	}
	time.Sleep(50 * time.Millisecond)
	if len(conn.MessagesOut) == conn.MessagesOutLimit {
		return &MockCloseError{}
	}
	conn.MessagesOut = append(conn.MessagesOut, data)
	return nil
}

func (conn *MockConn) Close() error {
	conn.IsClosed = true
	return nil
}
