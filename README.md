# dootroom

This repository is for experimenting with the WebSocket protocol and writing web applications in Go.

dootroom is a web application that broadcasts signals from one client to all other clients. It uses a similar architecture to the [Gorilla WebSocket's Chat example](https://github.com/gorilla/websocket/tree/master/examples/chat), except it handles connection-specific errors in a parent goroutine of the readPump and writePump goroutines.
