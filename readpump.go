package main

import "encoding/json"

// readPump runs a loop that reads a message from the connection, unmarshals
// JSON into a Diff, validates the Diff, and sends the Merge message to model.
func readPump(errSig *ErrorSignal, read Read, modelChan chan<- interface{}) {
	for {
		_, message, err := read()
		if err != nil {
			errSig.Close(err)
			return
		}
		diff := make(Diff)
		if err := json.Unmarshal(message, &diff); err != nil {
			errSig.Close(err)
			return
		}
		if err := validateDiff(diff); err != nil {
			errSig.Close(err)
			return
		}
		modelChan <- &Merge{diff}
	}
}
