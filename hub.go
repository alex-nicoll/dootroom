package main

type BufferOverflowError struct{}

func (err *BufferOverflowError) Error() string {
	return "Listener's buffer overflowed."
}

type Listener struct {
	sendChan chan<- []byte
	errSig   *ErrorSignal
}

// Register a Listener
type Register struct {
	li *Listener
}

// Unregister a Listener
type Unregister struct {
	li *Listener
}

// Broadcast a websocket message to all Listeners
type Broadcast struct {
	message []byte
}

// hub runs a loop that handles Register, Unregister, and Broadcast messages.
func hub(in <-chan interface{}) {
	listeners := make(map[*Listener]bool)
	for {
		switch m := (<-in).(type) {
		case *Register:
			listeners[m.li] = true
		case *Unregister:
			delete(listeners, m.li)
		case *Broadcast:
			for li := range listeners {
				select {
				case li.sendChan <- m.message:
				default:
					li.errSig.Close(&BufferOverflowError{})
					delete(listeners, li)
				}
			}
		}
	}
}
