package main

import (
	"errors"
	"testing"

	"dootroom.com/main/internal"
)

// readPump should close the error signal when the connection throws an error.
func Test_readPump(t *testing.T) {
	errSig := NewErrorSignal()
	conn := &internal.MockConn{In: make(chan []byte)}
	close(conn.In)
	unmarshalChan := make(chan interface{})

	readPump(errSig, conn, unmarshalChan)

	select {
	case <-errSig.Done():
		if _, ok := errSig.Err().(*internal.MockCloseError); !ok {
			t.Errorf("Closed the error signal, but with the wrong error")
		}
	default:
		t.Errorf("Didn't close the error signal")
	}

	return
}

// readPump should stop when the error signal is closed.
func Test_readPump2(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	conn := &internal.MockConn{In: make(chan []byte, 1)}
	conn.In <- []byte{}

	readPump(errSig, conn, nil)
}
