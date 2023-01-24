# multi-life

Conway's Game of Life implemented as a chaotic multiplayer game. Clients share and edit an ever-evolving grid.

[See it live](http://68.183.125.233/)

## Installation

1. Install [Docker Engine](https://docs.docker.com/engine/). If installing on Windows, use the WSL 2 backend.

2. Clone the repository.
```
git clone git@github.com:alex-nicoll/multi-life.git
```
3. `cd` into the repository's root directory and run `docker build` as shown below. This will build the application's Go backend from source inside a container, and output an executable named `server`.
```
cd multi-life
DOCKER_BUILDKIT=1 docker build . --output .
```
4. Run `server`, specifying a port number to listen on.
```
./server 8080
```
The application should now be running at http://localhost:8080 (or whichever port you specified).
