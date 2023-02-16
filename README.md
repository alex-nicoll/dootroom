# multi-life

Conway's Game of Life implemented as a chaotic multiplayer game. Clients share and edit an ever-evolving grid.

[See it live](https://multi-life-qngr9.ondigitalocean.app/)

## Installation

multi-life is available as a Docker image for linux/amd64.

1. Install [Docker Engine](https://docs.docker.com/engine/). If installing on Windows, use the WSL 2 backend.

2. Pull and run the image.
```
docker run -it --rm -p 8080:80 alexnicoll/multi-life
```
The application should now be running at http://localhost:8080.

3. Update the image with `docker pull alexnicoll/multi-life` as needed.

## Development

To develop multi-life, you will need Docker Engine and a POSIX shell. Use `build.sh` to build and test the application. You may specify a name and optional tag for the image (the default is multi-life:latest). E.g.,
```
./build.sh <name:tag>
```
To run the image,
```
docker run -it --rm -p 8080:80 multi-life
```
You can also bind-mount the assets directory to update the files being served without having to rebuild and restart the image. This is useful for making rapid changes to the client-side code.
```
docker run -it --rm -p 8080:80 -v <absolute-path-to-assets>:/assets multi-life
```
