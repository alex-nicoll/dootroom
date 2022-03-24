package main

import (
	"encoding/json"
	"errors"
)

type Unmarshal struct {
	errSig  *ErrorSignal
	message []byte
}

// unmarshal runs a loop that unmarshals data from the WebSocket message and
// sends the Merge message to model.
func unmarshal(in <-chan interface{}, modelChan chan<- interface{}) {
	for {
		m := (<-in).(*Unmarshal)
		diff := make(Diff)
		err := json.Unmarshal(m.message, &diff)
		if err != nil {
			m.errSig.Close(err)
		}
		for x := range diff {
			if x >= GridDimX {
				m.errSig.Close(errors.New("Unmarshalled diff exceeds grid's X dimension."))
			}
			for y := range diff[x] {
				if y >= GridDimY {
					m.errSig.Close(errors.New("Unmarshalled diff exceeds grid's Y dimension."))
				}
			}
		}
		modelChan <- &Merge{diff}
	}
}
