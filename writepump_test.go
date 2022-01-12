package main

import (
	"bytes"
	"testing"

	"dootroom.com/main/internal"
)

// writePump should copy messages to the websocket connection in order.
// writePump should call conn.Close() when the connection errors out.
func Test_writePump(t *testing.T) {
	messagesRecvd := [4][]byte{[]byte{0}, []byte{1}, []byte{2}, []byte{3}}
	send := make(chan []byte, 4)
	for _, m := range messagesRecvd {
		send <- m
	}
	conn := &internal.MockConn{MessagesOutLimit: 3}

	// Call writePump in this goroutine. It shouldn't block, because there are
	// 4 messages in the channel, and an error will be thrown after the first 3
	// messages are written to the connection.
	writePump(send, conn)

	for i := 0; i <= 2; i++ {
		if bytes.Compare(messagesRecvd[i], conn.MessagesOut[i]) != 0 {
			t.Errorf("message: Expected %v, got %v", messagesRecvd[i], conn.messagesOut[i])
		}
	}
	if len(conn.MessagesOut) > 3 {
		t.Errorf("messages written: Expected %v, got %v", 3, len(conn.MessagesOut))
	}
	if !conn.IsClosed {
		t.Errorf("connection closed: Expected %v, got %v", true, false)
	}
}

// writePump should call conn.Close() when the send channel is closed.
func Test_writePump2(t *testing.T) {
	// TODO
}
