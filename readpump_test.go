package main

import (
	"errors"
	"testing"
)

// readPump should stop when the error signal is closed.
func Test_readPump(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	read := func() (messageType int, p []byte, err error) {
		return 0, []byte{}, nil
	}

	readPump(errSig, read, nil)
}
