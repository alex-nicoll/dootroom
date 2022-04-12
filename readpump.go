package main

// readPump runs a loop that copies messages from the connection to unmarshalChan.
func readPump(errSig *ErrorSignal, read Read, unmarshalChan chan<- interface{}) {
	for {
		_, message, err := read()
		if err != nil {
			errSig.Close(err)
			return
		}
		unmarshalChan <- &Unmarshal{errSig, message}
	}
}
