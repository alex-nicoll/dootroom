// This script should be run after the HTML document has been loaded and
// parsed. It adds additional elements to the document, sets up event handlers,
// and sets up the WebSocket connection.

// Allow the modal to be opened and closed.
const modal = document.getElementById("modal_container");
document.getElementById("info").addEventListener("click", (e) => {
  modal.style.visibility = "visible";
});
document.getElementById("close").addEventListener("click", (e) => {
  modal.style.visibility = "hidden";
});

// Allow inputting the species (a seven-character hexadecimal color code).
let species = "#eaeaea";
const speciesInput = document.getElementById("species");
speciesInput.value = species;
speciesInput.addEventListener("input", (e) => {
  species = e.target.value;
});

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

// Map where the key is an overlay cell (div element) that has been filled, and
// the value is the species used to fill that cell.
const filledOverlayCells = new Map();

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
      cell.style.gridRow = `${x+1}`;
      cell.style.gridColumn = `${y+1}`;
      callback(cell, x, y);
      frag.appendChild(cell);
    }
  }
  return frag;
}

function fill(cell) {
  cell.className = "overlay_cell_filled";
  cell.style.backgroundColor = species;
  // Store the species along with the cell, to be sent to the server later. We
  // won't be able to use the value of style.backgroundColor, because it may be
  // converted from hexadecimal to something else (e.g., an RGB string),
  // whereas the server accepts only hexadecimal strings.
  filledOverlayCells.set(cell, species);
}

function empty(cell) {
  cell.className = "";
  cell.style.backgroundColor = "";
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
        const species = diff[x][y];
        if (species !== "") {
          cell.className = "grid_cell_filled";
          cell.style.backgroundColor = species;
        } else {
          cell.className = "grid_cell_empty";
          cell.style.backgroundColor = "";
        }
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
  filledOverlayCells.forEach((species, cell) => {
    const x_ysuffix = cell.id.split(",");
    const x = x_ysuffix[0];
    const y = x_ysuffix[1].split("-overlay")[0];
    if (diff[x] === undefined) {
      diff[x] = {};
    }
    diff[x][y] = species;

    empty(cell);
  });
  ws.send(JSON.stringify(diff));
}
