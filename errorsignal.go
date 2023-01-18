package main

import "sync"

// errorSignal is used to broadcast an error to multiple goroutines. Its
// methods can be called concurrently without further synchronization.
type errorSignal struct {
	rw  sync.RWMutex
	e   error
	sig chan struct{}
}

func newErrorSignal() *errorSignal {
	return &errorSignal{
		e:   nil,
		sig: make(chan struct{}),
	}
}

// signal returns a channel that, when closed, indicates that an error has been
// sent.
func (es *errorSignal) signal() <-chan struct{} {
	return es.sig
}

// err can be used to read the error once the channel returned by signal has
// been closed.
func (es *errorSignal) err() error {
	es.rw.RLock()
	defer es.rw.RUnlock()
	return es.e
}

// send closes the channel returned by signal and sets the error returned by
// err, effectively broadcasting an error to all goroutines receiving from the
// channel. Subsequent calls to send are no-ops.
func (es *errorSignal) send(err error) {
	if es.err() != nil {
		return
	}
	es.rw.Lock()
	defer es.rw.Unlock()
	es.e = err
	close(es.sig)
}
