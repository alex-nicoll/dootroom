# multi-life

Conway's Game of Life implemented as a chaotic multiplayer game. Clients share and edit an ever-evolving grid.

[See it live](http://68.183.125.233/)

## Installation

1. Install [Go](https://go.dev/doc/install).

2. Clone the repository.
```
git clone git@github.com:alex-nicoll/multi-life.git
```
3. `cd` into the repository's root directory and invoke `go run . <port_number>`, specifying the port on which you would like the server to run. E.g.:
```
cd multi-life
go run . 8080
```
The server should now be running at http://localhost:8080 (or whichever port you specified).

## Development

The following additional steps are needed to develop the project.

1. Install the third-party static analysis tools.
```
go install \
  github.com/mgechev/revive@latest \
  honnef.co/go/tools/cmd/staticcheck@latest \  
  github.com/kisielk/errcheck@latest
```
2. After making changes, use the `run` script to run static analysis, tests, and finally the application itself. The port number is optional and is 8080 by default.
```
./run <port_number>
```
