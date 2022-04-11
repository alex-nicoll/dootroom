package main

// Read decouples the application from Conn.ReadMessage for testing purposes.
type Read = func() (messageType int, p []byte, err error)

// Write decouples the application from Conn.WriteMessage for testing purposes.
type Write = func(messageType int, data []byte) error

// Close decouples the application from Conn.Close for testing purposes.
type Close = func() error
