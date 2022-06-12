// This script should be run after the HTML document has been loaded and
// parsed. It adds additional elements to the document, sets up event handlers,
// and sets up the WebSocket connection.
//
// You will see a few statements like this:
//   const x = (() => {...})();
// There will be some variables inside the ..., and some functions returned and
// assigned to x. The purpose is to keep the variables together with the
// functions that modify them. The expression on the right hand side of the
// assignment is often called an IIFE: Immediately Invoked Function Expression.
//
// There are also statments like this:
//   element.style.x = "something"
// This creates an inline style declaration for property x, overriding the
// style declared in CSS. Setting x to "" removes the inline style declaration,
// no longer overriding the style declared in CSS. See
// https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/style

// Allow icon buttons to change state when touched or moused over.
const iconButtons = document.getElementsByClassName("icon_button");
for (let i = 0; i < iconButtons.length; i++) {
  const ib = iconButtons.item(i);
  ib.addEventListener("pointerenter", (e) => {
    ib.classList.add("icon_button_over");
  });
  ib.addEventListener("pointerleave", (e) => {
    ib.classList.remove("icon_button_over");
  });
}

// Allow the modal to be opened and closed.
const modal = document.getElementById("modal_container");
iconButtons.namedItem("info").addEventListener("click", (e) => {
  modal.style.visibility = "";
});
iconButtons.namedItem("close").addEventListener("click", (e) => {
  modal.style.visibility = "hidden";
});

// Allow inputting the species (a seven-character hexadecimal color code).
// Start with a random species.
let species = "#" +
  Math.floor(Math.random() * Math.pow(2,24)).toString(16).padStart(6, "0");
const speciesInput = document.getElementById("species");
speciesInput.value = species;
speciesInput.addEventListener("input", (e) => {
  species = e.target.value;
});

// Create the grid cells.
const grid = document.getElementById("grid");
grid.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}`
  cell.className = "grid_cell_empty";
}));

// Create the overlay cells. 
const overlay_cells = document.getElementById("overlay_cells");
overlay_cells.appendChild(makeCells((cell, x, y) => {
  cell.id = `${x},${y}-overlay`
}));

// Prevent dragging of overlay cells.
overlay_cells.addEventListener("dragstart", (e) => {
  e.preventDefault();
});

// Prevent the board from scrolling when the mouse is pressed down inside the
// board and then dragged to the edge of the board.
const board = document.getElementById("board");
let isDraggingBoard;
board.addEventListener("mousedown", (e) => {
  isDraggingBoard = true;
});
board.addEventListener("mousemove", (e) => {
  e.preventDefault();
});
document.addEventListener("mouseup", (e) => {
  isDraggingBoard = false;
});

// Map where the key is an overlay cell (div element) that has been filled, and
// the value is the species used to fill that cell.
const filledOverlayCells = new Map();

// Object mouseDraw allows drawing and erasing by clicking or dragging with a
// mouse.
const mouseDraw = (() => {
  // drawState is either "drawing", "erasing", or undefined.
  let drawState;

  function handleMouseDown(e) {
    const cell = e.target;
    if (cell.className === "overlay_cell_filled") {
      empty(cell);
      drawState = "erasing";
    } else {
      fill(cell);
      drawState = "drawing";
    }
  }

  function handleMouseOver(e) {
    drawOrErase(drawState, e.target);
  }

  function handleMouseUp(e) {
    drawState = undefined;
  }

  return {
    enable: () => {
      overlay_cells.addEventListener("mousedown", handleMouseDown);
      overlay_cells.addEventListener("mouseover", handleMouseOver);
      document.addEventListener("mouseup", handleMouseUp);
    },
    disable: () => {
      overlay_cells.removeEventListener("mousedown", handleMouseDown);
      overlay_cells.removeEventListener("mouseover", handleMouseOver);
      document.removeEventListener("mouseup", handleMouseUp);
      drawState = undefined;
    }
  };
})();

// Object touchDraw allows drawing and erasing by dragging with a single touch.
// touchDraw only draws when a single touch moves. It doesn't draw when a
// single touch starts, in order to prevent accidental drawing in case of a
// multi-touch pan/zoom. As a result, we need some other way to handle taps.
const touchDraw = (() => {
  // drawState is either "drawing", "erasing", or undefined.
  let drawState;

  function handleTouchStart(e) {
    if (e.touches.length !== 1) {
      return;
    }
    if (e.target.className === "overlay_cell_filled") {
      drawState = "erasing";
    } else {
      drawState = "drawing";
    }
  }

  function handleTouchMove(e) {
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
      // Touch moved outside of overlay_cells.
      return;
    }
    drawOrErase(drawState, el);
  }

  function handleTouchEnd(e) {
    drawState = undefined;
  }

  function handleTouchCancel(e) {
    drawState = undefined;
  }

  return {
    enable: () => {
      overlay_cells.addEventListener("touchstart", handleTouchStart);
      overlay_cells.addEventListener("touchmove", handleTouchMove);
      document.addEventListener("touchend", handleTouchEnd);
      document.addEventListener("touchcancel", handleTouchCancel);
    },
    disable: () => {
      overlay_cells.removeEventListener("touchstart", handleTouchStart);
      overlay_cells.removeEventListener("touchmove", handleTouchMove);
      document.removeEventListener("touchend", handleTouchEnd);
      document.removeEventListener("touchcancel", handleTouchCancel);
      drawState = undefined;
    }
  };
})();

// Object tapDraw allows drawing and erasing by tapping with a single touch.
// You might ask, why is this object needed at all? The browser already fires
// mousedown when a tap ("click") is detected, so mouseDraw should handle taps.
// Well, on Safari for iOS and DuckDuckGo for Android, waiting for the
// mousedown event leads to a very obvious delay between tap and response.
const tapDraw = (() => {
  let isTapping = false;

  function handleTouchStart(e) {
    if (e.touches.length !== 1) {
      // Cancel the tap when multiple touches are detected.
      isTapping = false;
    }
    isTapping = true;
  }

  function handleTouchMove(e) {
    isTapping = false;
  }

  function handleTouchEnd(e) {
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
    // Prevent further events from firing, including mousedown (and mouseup,
    // and click). If mousedown were to fire with mouseDraw enabled, then we
    // would erase the cell that was just drawn.
    e.preventDefault();
  }

  function handleTouchCancel(e) {
    isTapping = false;
  }

  return {
    enable: () => {
      overlay_cells.addEventListener("touchstart", handleTouchStart);
      overlay_cells.addEventListener("touchmove", handleTouchMove);
      overlay_cells.addEventListener("touchend", handleTouchEnd);
      overlay_cells.addEventListener("touchcancel", handleTouchCancel);
    },
    disable: () => {
      overlay_cells.removeEventListener("touchstart", handleTouchStart);
      overlay_cells.removeEventListener("touchmove", handleTouchMove);
      overlay_cells.removeEventListener("touchend", handleTouchEnd);
      overlay_cells.removeEventListener("touchcancel", handleTouchCancel);
      isTapping = false;
    }
  };
})();

// Object mousePan allows panning by dragging with a mouse.
const mousePan = (() => {
  let isPanning = false;

  function handleMouseDown(e) {
    isPanning = true;
  }

  function handleMouseMove(e) {
    if (isPanning) {
      board.scrollTop -= e.movementY;
      board.scrollLeft -= e.movementX;
    }
  }

  function handleMouseUp(e) {
    isPanning = false;
  }

  return {
    enable: () => {
      board.addEventListener("mousedown", handleMouseDown);
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
    },
    disable: () => {
      board.removeEventListener("mousedown", handleMouseDown);
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      isPanning = false;
    }
  };
})();

// Allow switching between pan mode and draw/erase mode.
// isPanMode is true when we are in pan mode, and false when we are in
// draw/erase mode.
let isPanMode = false;
mouseDraw.enable();
touchDraw.enable();
tapDraw.enable();
const move = iconButtons.namedItem("move");
move.addEventListener("click", (e) => {
  if (isPanMode) {
    isPanMode = false;
    mousePan.disable();

    mouseDraw.enable();
    touchDraw.enable();
    tapDraw.enable();

    move.style.borderColor = "";

  } else {
    isPanMode = true;
    mousePan.enable();

    mouseDraw.disable();
    touchDraw.disable();
    tapDraw.disable();

    move.style.borderColor = "unset";
  }
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
iconButtons.namedItem("submit").addEventListener("click", submit);

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
