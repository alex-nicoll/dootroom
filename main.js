// Create the GoL grid.
const frag = document.createDocumentFragment();
for (let i = 0; i < 100; i++) {
  for (let j = 0; j < 100; j++) {
    const cell = document.createElement("div");
    cell.id = `${i},${j}`
    // CSS Grid rows and columns are indexed at 1, as opposed to 0.
    cell.style = `grid-row:${i+1};grid-column:${j+1};`;
    cell.className = "cell_empty";
    frag.appendChild(cell);
  }
}
const grid = document.getElementById("grid");
grid.appendChild(frag);

// Connect to the WebSocket server.
const ws = new WebSocket(`ws:\/\/${document.location.host}`);
ws.onmessage = (msg) => {
  msg.data.text().then((text) => {
    const diff = JSON.parse(text);
    console.log(diff);
    for (const i in diff) {
      for (const j in diff[i]) {
        const cell = document.getElementById(`${i},${j}`);
        cell.className = diff[i][j] ? "cell_filled" : "cell_empty";
      }
    }
  })
}
window.ws = ws;
