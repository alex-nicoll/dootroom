### Intro

This application implements a protocol on top of the WebSocket protocol. The protocol is designed to allow the client to make fast, evenly spaced out updates to its local Game of Life state.

All communication between the client and server is done in the form of text messages containing JSON.

If a client needs to close the WebSocket connection for any reason, it uses status code 1000 (normal closure).

### Grid

After the WebSocket connection is created, the server immediately sends a **grid**. This is a JSON array representing the entire Game of Life grid. Here is an example, where the Game of Life grid is 2x2:

`[["#aaaaaa",""],["#bbbbbb","#cccccc"]]`

A string element that is a hexadecimal color code (e.g. `"#aaaaaa"`) represents a live cell. An empty string element (`""`) represents a dead cell. No other elements may be included.

### Server Diff

After sending the grid, the server will begin sending **diff**s. A diff is a JSON object representing the difference between the current game state and the previous game state. A diff looks like so:

`{"0":{"0":"#dddddd","1":"#eeeeee"},"1":{"1":""}}`

A diff is indexed in the same way as a grid. I.e., in JavaScript, `JSON.parse(grid)[x][y]` and `JSON.parse(diff)[x][y]` refer to the same cell. Keys must be numeric strings in the range [0, dim), where dim is either the width or height in cells of the Game of Life grid.

The server sends diffs with an interval of approximately 170ms between them. The grid and first diff may be sent in quick succession.

The server sends a diff rather than a grid on each state change in order to reduce the amount of time the client spends updating its state, and reduce the amount of data sent over the network.

### Empty Diff

Messages from server to client can be considered to be broken up into a sequence of **stream**s. The current stream ends when the Game of Life grid has stopped evolving (due to it being empty or composed entirely of still lifes). When the stream ends, the server sends the empty diff:

`{}`

When evolution resumes, a new stream begins. The first stream on a connection consists of a grid followed by one or more diffs, and subsequent streams consist only of diffs.

The empty diff is used in order to simplify client-side buffering. A client may wish to smooth out irregularities in the rate of inbound diffs by buffering them and processing them at a regular interval, thereby creating the illusion of a "local" Game of Life. The empty diff tells the client that it can stop processing until another diff arrives.

### Client Diff

The client may send diffs representing changes to the game state. A client diff cannot contain `""` as an element and cannot be empty.

### Flow Control

If the rate of inbound diffs is too high for a client to process, the client may periodically reset the connection to get a new grid, at the cost of "skipping" the updates that would have occurred in between resets. The protocol may be improved in the future to include (1) allowing the client to request a grid rather than resetting the connection, and/or (2) slowing down the server to match the slowest maximum processing rate among clients.
