package main

// ConnReader decouples readPump from websocket.Conn for testing purposes.
type ConnReader interface {
	ReadMessage() (messageType int, p []byte, err error)
}

// readPump runs a loop that copies messages from conn to hubChan.
func readPump(errSig *ErrorSignal, conn ConnReader, hubChan chan<- interface{}) {
	for {
		_, message, err := conn.ReadMessage()
		select {
		case <-errSig.Done():
			return
		default:
			if err != nil {
				errSig.Close(err)
				return
			}
			hubChan <- &Broadcast{message}
		}
	}
}
