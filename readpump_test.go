package main

import (
	"errors"
	"testing"
)

type mockConnReader struct {
	in chan []byte
}

func (conn *mockConnReader) ReadMessage() (messageType int, p []byte, err error) {
	message, ok := <-conn.in
	if !ok {
		return 0, nil, errors.New("Client closed the connection.")
	}
	return 0, message, nil
}

// readPump should close the error signal when the connection throws an error.
func Test_readPump(t *testing.T) {
	errSig := NewErrorSignal()
	conn := &mockConnReader{in: make(chan []byte)}
	close(conn.in)
	unmarshalChan := make(chan interface{})

	readPump(errSig, conn, unmarshalChan)

	select {
	case <-errSig.Done():
	default:
		t.Errorf("Didn't close the error signal")
	}
}

// readPump should stop when the error signal is closed.
func Test_readPump2(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	conn := &mockConnReader{in: make(chan []byte, 1)}
	conn.in <- []byte{}

	readPump(errSig, conn, nil)
}
