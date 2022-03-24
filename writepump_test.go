package main

import (
	"errors"
	"testing"

	"dootroom.com/main/internal"
)

// writePump should close the error signal when the connection throws an error.
func Test_writePump(t *testing.T) {
	errSig := NewErrorSignal()
	sendChan := make(chan []byte, 2)
	sendChan <- []byte{}
	sendChan <- []byte{}
	conn := &internal.MockConn{Out: make(chan []byte, 1), OutLimit: 1}

	writePump(errSig, sendChan, conn)

	select {
	case <-errSig.Done():
		if _, ok := errSig.Err().(*internal.MockCloseError); !ok {
			t.Errorf("writePump closed the error signal, but with the wrong error")
		}
	default:
		t.Errorf("writePump didn't close the error signal")
	}

	return
}

// writePump should stop when the error signal is closed.
func Test_writePump2(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	sendChan := make(chan []byte)

	writePump(errSig, sendChan, nil)
}
