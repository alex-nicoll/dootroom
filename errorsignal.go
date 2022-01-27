package main

import "sync"

// ErrorSignal is used to broadcast an error to multiple goroutines. Its
// methods can be called concurrently without further synchronization.
type ErrorSignal struct {
	rw   sync.RWMutex
	err  error
	done chan struct{}
}

func NewErrorSignal() *ErrorSignal {
	return &ErrorSignal{
		err:  nil,
		done: make(chan struct{}),
	}
}

// Done returns a channel that, when closed, indicates that an error has
// occurred.
func (es *ErrorSignal) Done() <-chan struct{} {
	return es.done
}

// Err can be used to read the error once the channel returned by Done has been
// closed.
func (es *ErrorSignal) Err() error {
	es.rw.RLock()
	defer es.rw.RUnlock()
	return es.err
}

// Close closes the channel returned by Done and sets the error returned by
// Err, effectively broadcasting an error to all goroutines receiving from
// Done. Subsequent calls to Close are no-ops.
func (es *ErrorSignal) Close(err error) {
	if es.Err() != nil {
		return
	}
	es.rw.Lock()
	defer es.rw.Unlock()
	es.err = err
	close(es.done)
}
