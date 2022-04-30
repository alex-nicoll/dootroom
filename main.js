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

// Allow drawing and erasing by clicking or dragging with a mouse.
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

// Allow drawing and erasing by dragging with a single touch.
// touchDrawState is either "drawing", "erasing", or undefined.
let touchDrawState;
overlay.addEventListener("touchstart", (e) => {
  if (e.touches.length !== 1) {
    return;
  }
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
  touchDrawState = undefined;
});
document.addEventListener("touchcancel", (e) => {
  touchDrawState = undefined;
});

// Allow drawing and erasing by tapping with a single touch.
// OK Google, play One Touch by LCD Soundsystem.
// You might ask, why is this chunk of coded needed at all? The browser already
// fires mousedown and mouseup when a tap ("click") is detected. Well, on
// Safari for iOS and DuckDuckGo for Android, waiting for the mousedown event
// leads to a very obvious delay between tap and response.
let isTapping = false;
overlay.addEventListener("touchstart", (e) => {
  if (e.touches.length !== 1) {
    // Cancel the tap when multiple touches are detected.
    isTapping = false;
  }
  isTapping = true;
});
overlay.addEventListener("touchmove", (e) => {
  isTapping = false;
});
overlay.addEventListener("touchend", (e) => {
  if (e.touches.length !== 0) {
    // There are still touches on the touch surface, so this isn't a tap.
    isTapping = false;
    return;
  }
  if (!isTapping) {
    // This is the end of a one-touch movement, or a multi-touch interaction.
    return;
  }
  const cell = e.target;
  if (cell.className === "overlay_cell_filled") {
    empty(cell);
  } else {
    fill(cell);
  }
  isTapping = false;
  // Prevent further events from firing (including mousedown and mouseup).
  e.preventDefault();
});
overlay.addEventListener("touchcancel", (e) => {
  isTapping = false;
});

// Connect to the WebSocket server.
let ws = connect();

// Disconnect when the page is hidden, and reconnect when it's visible again.
document.addEventListener("visibilitychange", (e) => {
  if (document.visibilityState === "hidden") {
    ws.close(1000, "page hidden");
  } else if (document.visibilityState === "visible") {
    ws = connect();
  }
});

// Allow submitting via the Enter key.
document.addEventListener("keydown", (e) => {
  if (e.code === "Enter") {
    submit();
  }
});

// Allow submitting via the submit button.
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

function connect() {
  const ws = new WebSocket(`ws:\/\/${document.location.host}`);
  ws.addEventListener("message", update);
  return ws;
}

// update handles a grid change coming from the server.
function update(msg) {
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
