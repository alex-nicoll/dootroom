package main

import (
	"bytes"
	"sync"
	"testing"

	"dootroom.com/main/internal"
)

// readPump should copy messages to hubChan in order.
// readPump should close the error signal when the connection errors out.
func Test_readPump(t *testing.T) {
	errSig := NewErrorSignal()
	messagesIn := [3][]byte{[]byte{0}, []byte{1}, []byte{2}}
	conn := &internal.MockConn{MessagesIn: messagesIn[:]}
	hubChan := make(chan interface{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		readPump(errSig, conn, hubChan)
	}()

	messagesHubChan := [3]interface{}{<-hubChan, <-hubChan, <-hubChan}
	for i := 0; i <= 2; i++ {
		message := messagesHubChan[i].(*Broadcast).message
		if bytes.Compare(messagesIn[i], message) != 0 {
			t.Errorf("message: Expected %v, got %v", messagesIn[i], message)
		}
	}

	wg.Wait()

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
	// TODO
}
