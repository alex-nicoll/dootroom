# mGoL

Conway's Game of Life implemented as a chaotic multiplayer game. Clients share the grid and can modify it while it evolves.

[See it live](http://68.183.125.233/)

## Installation

1. Install [Go](https://go.dev/doc/install).

2. Clone the repository.
```
git clone git@github.com:alex-nicoll/mGoL.git
```
3. `cd` into the repository's root directory, and invoke `run`. The port number is optional and is 8080 by default.
```
cd mGoL
./run <port_number>
```
The server should now be running at http://localhost:8080 (or whichever port you specified).
