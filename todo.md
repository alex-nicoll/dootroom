- Simplify model. Server only needs to send the cells that changed state, not the new state. Likewise for the client, unless we implement erasing.
- Write test for client side code (input: ws; output: document)
- Generate server and client code so that the grid dimensions are defined in one place
- Do the tests leak goroutines?
- For tests in which blocking forever indicates failure, add timeouts to help with debugging
- Pick an appropriate port number.
- Pick appropriate WebSocket buffer sizes to pass to Upgrader. See Gorilla WebSocket documentation.
- Look into using Gorilla WebSocket readJSON and writeJSON methods
- Docker it up so app can run on any OS
- Cloud infrastructure
- CI/CD pipeline
- Browser testing with Selenium (or BrowserStack/LambdaTest to hit macOS)
- Design doc

Features:

- Draw on teal overlay, submit
- Stamps
- Auto reset, followed by free draw period
- Stamp builder
- Langton's Ant
- Gamepad support
- Something that requires a database
