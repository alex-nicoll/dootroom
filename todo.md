- Link to github and game of life wiki
- Max clients
- Handle server sending updates faster than UI can render them.
  - Have client ask for another init when too many diffs are buffered.
  - Profile and speed up rendering.
- Reduce overloading of the term "grid". Currently refers to CSS module, HTML element, and backend data structure.
- Try using SVG or canvas instead of HTML div's for grid
- On init, server can send just the live cells, rather than the whole grid, so long as the client can distinguish between an init and a diff.
- Write test for client side code (input: ws; output: document)
- Generate server and client code so that the grid dimensions are defined in one place
- Do the tests leak goroutines?
- Pick appropriate WebSocket buffer sizes to pass to Upgrader. See Gorilla WebSocket documentation.
- Look into using Gorilla WebSocket readJSON and writeJSON methods
- Docker it up so app can run on any OS
- Browser testing with Selenium (or BrowserStack/LambdaTest to hit macOS)
- CI/CD pipeline
- Design doc. This would be a good place to explain that gol+hub+writePump maintains the order of InitListener and Tick messages, so that the client stays in sync with the server.

Features:

- Allow grid to zoom independently of other content (iframe)
- A way to zoom on devices that don't have pinch (e.g. mouse only)
- Toroidal array
- Auto reset, followed by free draw period
- Langton's Ant
- Stamps
- Stamp builder
- Gamepad support
- Something that requires a database
