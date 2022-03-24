package internal

type MockCloseError struct{}

func (err *MockCloseError) Error() string {
	return "Client closed the connection."
}

type MockConn struct {
	In       chan []byte
	Out      chan []byte
	outCount int
	// If OutLimit is an int, the (OutLimit+1)-th call to WriteMessage will return an error
	// If OutLimit is nil, WriteMessage will never return an error
	OutLimit interface{}
	IsClosed bool
}

func (conn *MockConn) ReadMessage() (messageType int, p []byte, err error) {
	message, ok := <-conn.In
	if !ok {
		return 0, nil, &MockCloseError{}
	}
	return 0, message, nil
}

func (conn *MockConn) WriteMessage(messageType int, data []byte) error {
	if lim, ok := conn.OutLimit.(int); ok {
		if conn.outCount == lim {
			return &MockCloseError{}
		}
		conn.outCount++
	}
	conn.Out <- data
	return nil
}

func (conn *MockConn) Close() error {
	conn.IsClosed = true
	return nil
}
