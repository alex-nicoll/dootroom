// This script should be run after the HTML document has been loaded and
// parsed. It adds additional elements to the document, sets up event handlers,
// and sets up the WebSocket connection.

// Create the grid.
const grid = document.getElementById("grid");
grid.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}`
  cell.className = "grid_cell_empty";
}));

// Set of overlay cells (div elements) that have been drawn (filled in by clicking or dragging over).
const drawnOverlayCells = new Set();

// drawState is either "drawing", "erasing", or undefined.
let drawState;

// Create the overlay. Handle pointer events to allow drawing and erasing cells.
const overlay = document.getElementById("overlay");
overlay.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}-overlay`
  cell.addEventListener("pointerdown", (e) => {
    const cell = e.target;
    if (cell.className === "overlay_cell_filled") {
      erase(cell);
      drawState = "erasing";
    } else {
      draw(cell);
      drawState = "drawing";
    }
  });
  cell.addEventListener("pointerenter", (e) => {
    const cell = e.target;
    if (drawState === "drawing" && cell.className === "") {
      draw(cell);
    } else if (drawState === "erasing" && cell.className === "overlay_cell_filled") {
      erase(cell);
    }
  });
}));
document.addEventListener("pointerup", (e) => drawState = undefined);

// Disable dragging of overlay elements.
overlay.addEventListener("dragstart", (e) => e.preventDefault());

// Connect to the WebSocket server.
const ws = new WebSocket(`ws:\/\/${document.location.host}`);

// Handle changes to the grid coming from the server.
ws.onmessage = (msg) => {
  msg.data.text().then((text) => {
    const diff = JSON.parse(text);
    for (const x in diff) {
      for (const y in diff[x]) {
        const cell = document.getElementById(`${x},${y}`);
        cell.className = diff[x][y] ? "grid_cell_filled" : "grid_cell_empty";
      }
    }
  });
}

// Allow the grid changes represented by the drawn overlay cells to be
// submitted to the server via the Enter key.
document.addEventListener("keydown", (e) => {
  if (e.code === "Enter" && drawnOverlayCells.size !== 0) {
    const diff = {};
    for (const cell of drawnOverlayCells) {
      const x_ysuffix = cell.id.split(",");
      const x = x_ysuffix[0];
      const y = x_ysuffix[1].split("-overlay")[0];
      if (diff[x] === undefined) {
        diff[x] = {};
      }
      diff[x][y] = true;

      erase(cell);
    }
    ws.send(JSON.stringify(diff));
  }
});

function makeCells(callback) {
  const frag = document.createDocumentFragment();
  for (let x = 0; x < 100; x++) {
    for (let y = 0; y < 100; y++) {
      const cell = document.createElement("div");
      // CSS Grid rows and columns are indexed at 1, as opposed to 0.
      cell.style = `grid-row:${x+1};grid-column:${y+1};`;
      callback(cell, x, y);
      frag.appendChild(cell);
    }
  }
  return frag;
}

function draw(cell) {
  cell.className = "overlay_cell_filled";
  drawnOverlayCells.add(cell);
}

function erase(cell) {
  cell.className = "";
  drawnOverlayCells.delete(cell);
}
