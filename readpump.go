package main

// ConnReader decouples readPump from websocket.Conn for testing purposes.
type ConnReader interface {
	ReadMessage() (messageType int, p []byte, err error)
}

// readPump runs a loop that copies messages from conn to unmarshalChan.
func readPump(errSig *ErrorSignal, conn ConnReader, unmarshalChan chan<- interface{}) {
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
			unmarshalChan <- &Unmarshal{errSig, message}
		}
	}
}
