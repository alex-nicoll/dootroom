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
		if err := json.Unmarshal(m.message, &diff); err != nil {
			m.errSig.Close(err)
			continue
		}
		if err := validateDiff(diff); err != nil {
			m.errSig.Close(err)
			continue
		}
		modelChan <- &Merge{diff}
	}
}

func validateDiff(diff Diff) error {
	if len(diff) == 0 {
		return errors.New("Unmarshalled diff is empty.")
	}
	for x := range diff {
		if x >= GridDimX {
			return errors.New("Unmarshalled diff exceeds grid's X dimension.")
		}
		if len(diff[x]) == 0 {
			return errors.New("Unmarshalled diff includes an X coordinate with no Y coordinate.")
		}
		for y := range diff[x] {
			if y >= GridDimY {
				return errors.New("Unmarshalled diff exceeds grid's Y dimension.")
			}
		}
	}
	return nil
}
