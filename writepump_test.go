package main

import (
	"errors"
	"testing"
)

// writePump should close the error signal when the connection throws an error.
func Test_writePump(t *testing.T) {
	errSig := NewErrorSignal()
	sendChan := make(chan []byte, 1)
	sendChan <- []byte{}
	write := func(messageType int, data []byte) error {
		return errors.New("Client closed the connection.")
	}

	writePump(errSig, func(err error) {}, sendChan, write)

	select {
	case <-errSig.Done():
	default:
		t.Errorf("writePump didn't close the error signal")
	}
}
