package main

type bufferOverflowError struct{}

func (err *bufferOverflowError) Error() string {
	return "Listener's buffer overflowed."
}

type listener struct {
	sendChan chan<- []byte
	errSig   *errorSignal
}
