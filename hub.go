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

// Broadcast a websocket message to all registered Listeners
type Broadcast struct {
	message []byte
}

// Forward a websocket message to a Specific listener
type Forward struct {
	li      *Listener
	message []byte
}

// hub runs a loop that sends websocket messages to Listeners.
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
		case *Forward:
			select {
			case m.li.sendChan <- m.message:
			default:
				m.li.errSig.Close(&BufferOverflowError{})
				delete(listeners, m.li)
			}
		}
	}
}
