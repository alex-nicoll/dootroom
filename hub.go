package main

type Hub struct {
	// Register requests.
	register chan MsgChan

	// Unregister requests.
	unregister chan MsgChan

	// Inbound messages.
	broadcast MsgChan

	// Registered outbound channels.
	sends map[MsgChan]bool
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan MsgChan),
		unregister: make(chan MsgChan),
		broadcast:  make(MsgChan),
		sends:      make(map[MsgChan]bool),
	}
}

// run broadcasts messages to the set of active outbound channels,
// and handles requests to register or unregister outbound channels.
func (hub *Hub) run() {
	for {
		select {
		case send := <-hub.register:
			hub.sends[send] = true
		case send := <-hub.unregister:
			if _, ok := hub.sends[send]; ok {
				delete(hub.sends, send)
				close(send)
			}
		case message := <-hub.broadcast:
			for send := range hub.sends {
				select {
				case send <- message:
				default:
					close(send)
					delete(hub.sends, send)
					// TODO: Think about this a bit more.
					// Is this the right thing to do when client's send buffer is full?
				}
			}
		}
	}
}
