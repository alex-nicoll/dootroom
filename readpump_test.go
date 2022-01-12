package main

import (
	"bytes"
	"sync"
	"testing"

	"dootroom.com/main/internal"
)

// readPump should copy messages to the broadcast channel in order.
// readPump should call conn.Close() and unregister() when the connection errors out.
func Test_readPump(t *testing.T) {
	messagesIn := [3][]byte{[]byte{0}, []byte{1}, []byte{2}}
	conn := &internal.MockConn{MessagesIn: messagesIn[:]}
	broadcast := make(chan []byte)
	didUnregister := false

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		readPump(conn, broadcast, func() { didUnregister = true })
	}()

	messagesBcast := [3][]byte{<-broadcast, <-broadcast, <-broadcast}
	for i := 0; i <= 2; i++ {
		if bytes.Compare(messagesIn[i], messagesBcast[i]) != 0 {
			t.Errorf("message: Expected %v, got %v", messagesIn[i], messagesBcast[i])
		}
	}

	wg.Wait()
	if !conn.IsClosed {
		t.Errorf("connection closed: Expected %v, got %v", true, false)
	}
	if !didUnregister {
		t.Errorf("unregistered: Expected %v, got %v", true, false)
	}
	return
}
