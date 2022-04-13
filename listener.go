package main

type BufferOverflowError struct{}

func (err *BufferOverflowError) Error() string {
	return "Listener's buffer overflowed."
}

type Listener struct {
	sendChan chan<- []byte
	errSig   *ErrorSignal
}
