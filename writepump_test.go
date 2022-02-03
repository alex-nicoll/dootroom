package main

import (
	"bytes"
	"errors"
	"testing"

	"dootroom.com/main/internal"
)

// writePump should copy messages to the websocket connection in order.
// writePump should close the error signal when the connection errors out.
func Test_writePump(t *testing.T) {
	errSig := NewErrorSignal()
	messagesRecvd := [4][]byte{[]byte{0}, []byte{1}, []byte{2}, []byte{3}}
	sendChan := make(chan []byte, 4)
	for _, m := range messagesRecvd {
		sendChan <- m
	}
	conn := &internal.MockConn{MessagesOutLimit: 3}

	// Call writePump in this goroutine. It shouldn't block, because there are
	// 4 messages in the channel, and an error will be thrown after the first 3
	// messages are written to the connection.
	writePump(errSig, sendChan, conn)

	for i := 0; i <= 2; i++ {
		if bytes.Compare(messagesRecvd[i], conn.MessagesOut[i]) != 0 {
			t.Errorf("message: Expected %v, got %v", messagesRecvd[i], conn.MessagesOut[i])
		}
	}

	select {
	case <-errSig.Done():
		if _, ok := errSig.Err().(*internal.MockCloseError); !ok {
			t.Errorf("readPump closed the error signal, but with the wrong error")
		}
	default:
		t.Errorf("readPump didn't close the error signal")
	}

	return
}

// writePump should stop when the error signal is closed.
func Test_writePump2(t *testing.T) {
	errSig := NewErrorSignal()
	errSig.Close(errors.New("dummy error"))
	sendChan := make(chan []byte, 1)
	sendChan <- []byte{}
	conn := &internal.MockConn{MessagesOutLimit: 2}

	writePump(errSig, sendChan, conn)
}
