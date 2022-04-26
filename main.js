// This script should be run after the HTML document has been loaded and
// parsed. It adds additional elements to the document, sets up event handlers,
// and sets up the WebSocket connection.

// Create the grid.
const grid = document.getElementById("grid");
grid.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}`
  cell.className = "grid_cell_empty";
}));

// Create the overlay. 
const overlay = document.getElementById("overlay");
overlay.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}-overlay`
}));

// Prevent dragging of overlay elements.
overlay.addEventListener("dragstart", (e) => {
  e.preventDefault();
});

// Set of overlay cells (div elements) that have been filled.
const filledOverlayCells = new Set();

// Allow drawing and erasing with a mouse.
// mouseDrawState is either "drawing", "erasing", or undefined.
let mouseDrawState;
overlay.addEventListener("mousedown", (e) => {
  const cell = e.target;
  if (cell.className === "overlay_cell_filled") {
    empty(cell);
    mouseDrawState = "erasing";
  } else {
    fill(cell);
    mouseDrawState = "drawing";
  }
});
overlay.addEventListener("mouseover", (e) => {
  drawOrErase(mouseDrawState, e.target);
});
document.addEventListener("mouseup", (e) => {
  mouseDrawState = undefined;
});

// Allow drawing and erasing with a single touch. User can tap on cells to
// change them, or move one finger across the screen to draw/erase. Tapping
// with two fingers does nothing. Moving two fingers across the screen should
// invoke the browser's default behavior - scrolling, hopefully.
// touchDrawState is either "drawing", "erasing", or undefined.
let touchDrawState;
let isTapping = false;
overlay.addEventListener("touchstart", (e) => {
  if (e.touches.length !== 1) {
    // Stop registering a tap as soon as a second touch is detected.
    isTapping = false;
    return;
  }
  isTapping = true;
  if (e.target.className === "overlay_cell_filled") {
    touchDrawState = "erasing";
  } else {
    touchDrawState = "drawing";
  }
});
overlay.addEventListener("touchmove", (e) => {
  if (e.touches.length !== 1) {
    return;
  }
  isTapping = false;
  if (e.cancelable) {
    // Prevent scrolling.
    e.preventDefault();
  }
  const touch = e.touches.item(0);
  const el = document.elementFromPoint(touch.clientX, touch.clientY);
  if (!el) {
    // Touch moved outside of the viewport.
    return;
  }
  if (!el.id.endsWith("-overlay")) {
    // Touch moved outside of the overlay.
    return;
  }
  drawOrErase(touchDrawState, el);
});
document.addEventListener("touchend", (e) => {
  if (isTapping) {
    drawOrErase(touchDrawState, e.target);
    isTapping = false;
    // Prevent further events from firing (including mousedown and mouseup).
    e.preventDefault();
  }
  touchDrawState = undefined;
});
document.addEventListener("touchcancel", (e) => {
  isTapping = false;
  touchDrawState = undefined;
});

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

// Allow submitting via the Enter key.
document.addEventListener("keydown", (e) => {
  if (e.code === "Enter") {
    submit();
  }
});

// Allow submitting via submit button.
document.getElementById("submit").addEventListener("click", submit);

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

function fill(cell) {
  cell.className = "overlay_cell_filled";
  filledOverlayCells.add(cell);
}

function empty(cell) {
  cell.className = "";
  filledOverlayCells.delete(cell);
}

function drawOrErase(drawState, cell) {
  if (drawState === "drawing" && cell.className === "") {
    fill(cell);
  } else if (drawState === "erasing" && cell.className === "overlay_cell_filled") {
    empty(cell);
  }
}

// submit sends the grid changes represented by the filled overlay cells to the server.
function submit() {
  if (filledOverlayCells.size === 0) {
    return;
  }
  const diff = {};
  for (const cell of filledOverlayCells) {
    const x_ysuffix = cell.id.split(",");
    const x = x_ysuffix[0];
    const y = x_ysuffix[1].split("-overlay")[0];
    if (diff[x] === undefined) {
      diff[x] = {};
    }
    diff[x][y] = true;

    empty(cell);
  }
  ws.send(JSON.stringify(diff));
}
