// This script should be run after the HTML document has been loaded and
// parsed. It adds additional elements to the document, sets up the WebSocket
// connection, and sets up event handlers.

// Create the grid.
const grid = document.getElementById("grid");
grid.appendChild(makeCells("", "grid_cell_empty"));

// Create the overlay.
const overlay = document.getElementById("overlay");
overlay.appendChild(makeCells("-overlay", ""));

function makeCells(idSuffix, className) {
  const frag = document.createDocumentFragment();
  for (let x = 0; x < 100; x++) {
    for (let y = 0; y < 100; y++) {
      const cell = document.createElement("div");
      cell.id = `${x},${y}${idSuffix}`
      // CSS Grid rows and columns are indexed at 1, as opposed to 0.
      cell.style = `grid-row:${x+1};grid-column:${y+1};`;
      cell.className = className;
      frag.appendChild(cell);
    }
  }
  return frag;
}

// Connect to the WebSocket server.
const ws = new WebSocket(`ws:\/\/${document.location.host}`);

// Handle changes to the grid coming from the server.
ws.onmessage = (msg) => {
  msg.data.text().then((text) => {
    const diff = JSON.parse(text);
    console.log(diff);
    for (const x in diff) {
      for (const y in diff[x]) {
        const cell = document.getElementById(`${x},${y}`);
        cell.className = diff[x][y] ? "grid_cell_filled" : "grid_cell_empty";
      }
    }
  });
}

const selectedCells = new Set();

// Allow overlay cells to become selected.
overlay.addEventListener("mousedown", (e) => {
  const cell = e.target;
  if (cell.className === "overlay_cell_filled") {
    cell.className = "";
    selectedCells.delete(cell);
  } else {
    cell.className = "overlay_cell_filled";
    selectedCells.add(cell);
  }
});

// Allow the changes represented by the selected cells to be submitted to the
// server via the Enter key.
document.addEventListener("keydown", (e) => {
  if (e.code === "Enter" && selectedCells.size !== 0) {
    const diff = {};
    for (const cell of selectedCells) {
      const x_ysuffix = cell.id.split(",");
      const x = x_ysuffix[0];
      const y = x_ysuffix[1].split("-overlay")[0];
      if (diff[x] === undefined) {
        diff[x] = {};
      }
      diff[x][y] = true;

      cell.className = "";
      selectedCells.delete(cell);
    }
    ws.send(JSON.stringify(diff));
  }
});
