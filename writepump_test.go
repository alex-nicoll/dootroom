package main

import (
	"errors"
	"testing"
)

type mockConnWriter struct {
	out      chan []byte
	outCount int
	outLimit int
}

func (conn *mockConnWriter) WriteMessage(messageType int, data []byte) error {
	if conn.outCount == conn.outLimit {
		return errors.New("Client closed the connection.")
	}
	conn.outCount++
	conn.out <- data
	return nil
}

// writePump should close the error signal when the connection throws an error.
func Test_writePump(t *testing.T) {
	errSig := NewErrorSignal()
	sendChan := make(chan []byte, 2)
	sendChan <- []byte{}
	sendChan <- []byte{}
	conn := &mockConnWriter{out: make(chan []byte, 1), outLimit: 1}

	writePump(errSig, sendChan, conn)

	select {
	case <-errSig.Done():
	default:
		t.Errorf("writePump didn't close the error signal")
	}
}

// writePump should stop when the error signal is closed.
func Test_writePump2(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	sendChan := make(chan []byte)

	writePump(errSig, sendChan, nil)
}
