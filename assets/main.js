// This script should be included as a module so that it runs after the HTML
// document has been loaded and parsed. It adds additional elements to the
// document, sets up event handlers, and sets up the WebSocket connection.
//
// You will see a few statments like this:
//   element.style.x = "something";
// This creates an inline style declaration for property x, overriding the
// style declared in CSS. Setting x to "" removes the inline style declaration,
// no longer overriding the style declared in CSS. See
// https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/style

init();

function init() {

  initBoardCells();

  const overlayCells = document.getElementById("overlay_cells");
  overlayCells.appendChild(makeCells((cell, x, y) => {
    cell.id = `${x},${y}-overlay`;
  }));
  // Prevent dragging of overlay cells.
  overlayCells.addEventListener("dragstart", (e) => {
    e.preventDefault();
  });

  const iconButtons = document.getElementsByClassName("icon_button");
  // Allow icon buttons to change state when touched or moused over.
  for (let i = 0; i < iconButtons.length; i++) {
    const ib = iconButtons.item(i);
    ib.addEventListener("pointerenter", () => {
      ib.classList.add("icon_button_over");
    });
    ib.addEventListener("pointerleave", () => {
      ib.classList.remove("icon_button_over");
    });
  }
  initModal(iconButtons);

  const speciesInput = document.getElementById("species");

  // The species is a seven-character hexadecimal color code. Start with a
  // random value.
  const initialSpecies = "#" + Math.floor(Math.random() *
    Math.pow(2,24)).toString(16).padStart(6, "0")
  speciesInput.value = initialSpecies;

  // filledOverlayCells is a map where the key is an overlay cell that has been
  // filled, and the value is the species used to fill that cell.
  const filledOverlayCells = new Map();

  const mouseDraw = newMouseDraw(overlayCells, filledOverlayCells,
    speciesInput);
  const touchDraw = newTouchDraw(overlayCells, filledOverlayCells,
    speciesInput);
  const tapDraw = newTapDraw(overlayCells, filledOverlayCells, speciesInput);

  const view = document.getElementById("view");
  initView(view);
  const mousePan = newMousePan(view);

  initModeSwitch(iconButtons, mouseDraw, touchDraw, tapDraw, mousePan);

  const processor = newProcessor();
  const ws = newWs(processor, filledOverlayCells);
  const balancer = newBalancer(ws, processor);

  // Allow submitting via the submit button.
  iconButtons.namedItem("submit").addEventListener("click", ws.submit);
  // Allow submitting via the Enter key.
  document.addEventListener("keydown", (e) => {
    if (e.code === "Enter") {
      ws.submit();
    }
  });
  initVisChangeHandling(ws, balancer);

  balancer.start();
  // Use setTimeout to ensure that all the costly DOM updates in this script
  // complete before we start receiving messages to process.
  setTimeout(ws.connect, 0);
}

function fill(filledOverlayCells, cell, species) {
  cell.className = "overlay_cell_filled";
  cell.style.backgroundColor = species;
  // Store the species along with the cell, to be sent to the server later. We
  // won't be able to use the value of style.backgroundColor, because it may be
  // converted from hexadecimal to something else (e.g., an RGB string),
  // whereas the server accepts only hexadecimal strings.
  filledOverlayCells.set(cell, species);
}

function empty(filledOverlayCells, cell) {
  cell.className = "";
  cell.style.backgroundColor = "";
  filledOverlayCells.delete(cell);
}

// Flush empties all of the filled overlay cells and converts them to a diff to
// submit to the server.
function flush(filledOverlayCells) {
  const diff = {};
  filledOverlayCells.forEach((species, cell) => {
    const x_ysuffix = cell.id.split(",");
    const x = x_ysuffix[0];
    const y = x_ysuffix[1].split("-overlay")[0];
    if (diff[x] === undefined) {
      diff[x] = {};
    }
    diff[x][y] = species;

    empty(filledOverlayCells, cell);
  });
  return diff;
}

function drawOrErase(drawState, filledOverlayCells, cell, species) {
  if (drawState === "drawing" && cell.className === "") {
    fill(filledOverlayCells, cell, species);
  } else if (drawState === "erasing" && cell.className === "overlay_cell_filled") {
    empty(filledOverlayCells, cell);
  }
}

// MouseDraw allows drawing and erasing by clicking or dragging with a mouse.
function newMouseDraw(overlayCells, filledOverlayCells, speciesInput) {
  // drawState is either "drawing", "erasing", or undefined.
  let drawState;

  function handleMouseDown(e) {
    const cell = e.target;
    if (cell.className === "overlay_cell_filled") {
      empty(filledOverlayCells, cell);
      drawState = "erasing";
    } else {
      fill(filledOverlayCells, cell, speciesInput.value);
      drawState = "drawing";
    }
  }

  function handleMouseOver(e) {
    drawOrErase(drawState, filledOverlayCells, e.target, speciesInput.value);
  }

  function handleMouseUp() {
    drawState = undefined;
  }

  function enable() {
    overlayCells.addEventListener("mousedown", handleMouseDown);
    overlayCells.addEventListener("mouseover", handleMouseOver);
    document.addEventListener("mouseup", handleMouseUp);
  }

  function disable() {
    overlayCells.removeEventListener("mousedown", handleMouseDown);
    overlayCells.removeEventListener("mouseover", handleMouseOver);
    document.removeEventListener("mouseup", handleMouseUp);
    drawState = undefined;
  }

  return { enable, disable };
}

// TouchDraw allows drawing and erasing by dragging with a single touch.
// TouchDraw only draws when a single touch moves. It doesn't draw when a
// single touch starts, in order to prevent accidental drawing in case of a
// multi-touch pan/zoom. As a result, we need some other way to handle taps.
function newTouchDraw(overlayCells, filledOverlayCells, speciesInput) {
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
    drawOrErase(drawState, filledOverlayCells, el, speciesInput.value);
  }

  function handleTouchEnd() {
    drawState = undefined;
  }

  function handleTouchCancel() {
    drawState = undefined;
  }

  function enable() {
    overlayCells.addEventListener("touchstart", handleTouchStart);
    overlayCells.addEventListener("touchmove", handleTouchMove);
    document.addEventListener("touchend", handleTouchEnd);
    document.addEventListener("touchcancel", handleTouchCancel);
  }

  function disable() {
    overlayCells.removeEventListener("touchstart", handleTouchStart);
    overlayCells.removeEventListener("touchmove", handleTouchMove);
    document.removeEventListener("touchend", handleTouchEnd);
    document.removeEventListener("touchcancel", handleTouchCancel);
    drawState = undefined;
  }

  return { enable, disable };
}

// TapDraw allows drawing and erasing by tapping with a single touch. You might
// ask, why is this object needed at all? The browser already fires mousedown
// when a tap ("click") is detected, so MouseDraw should handle taps. Well, on
// Safari for iOS and DuckDuckGo for Android, waiting for the mousedown event
// leads to a very obvious delay between tap and response.
function newTapDraw(overlayCells, filledOverlayCells, speciesInput) {
  let isTapping = false;

  function handleTouchStart(e) {
    if (e.touches.length !== 1) {
      // Cancel the tap when multiple touches are detected.
      isTapping = false;
    }
    isTapping = true;
  }

  function handleTouchMove() {
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
      empty(filledOverlayCells, cell);
    } else {
      fill(filledOverlayCells, cell, speciesInput.value);
    }
    isTapping = false;
    // Prevent further events from firing, including mousedown (and mouseup,
    // and click). If mousedown were to fire with mouseDraw enabled, then we
    // would erase the cell that was just drawn.
    e.preventDefault();
  }

  function handleTouchCancel() {
    isTapping = false;
  }

  function enable() {
    overlayCells.addEventListener("touchstart", handleTouchStart);
    overlayCells.addEventListener("touchmove", handleTouchMove);
    overlayCells.addEventListener("touchend", handleTouchEnd);
    overlayCells.addEventListener("touchcancel", handleTouchCancel);
  }

  function disable() {
    overlayCells.removeEventListener("touchstart", handleTouchStart);
    overlayCells.removeEventListener("touchmove", handleTouchMove);
    overlayCells.removeEventListener("touchend", handleTouchEnd);
    overlayCells.removeEventListener("touchcancel", handleTouchCancel);
    isTapping = false;
  }

  return { enable, disable };
}

// MousePan allows panning by dragging with a mouse.
function newMousePan(view) {
  let isPanning = false;

  function handleMouseDown() {
    isPanning = true;
  }

  function handleMouseMove(e) {
    if (isPanning) {
      view.scrollTop -= e.movementY;
      view.scrollLeft -= e.movementX;
    }
  }

  function handleMouseUp() {
    isPanning = false;
  }

  function enable() {
    view.addEventListener("mousedown", handleMouseDown);
    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
  }

  function disable() {
    view.removeEventListener("mousedown", handleMouseDown);
    document.removeEventListener("mousemove", handleMouseMove);
    document.removeEventListener("mouseup", handleMouseUp);
    isPanning = false;
  }

  return { enable, disable };
}

// Ws supports connecting and disconnecting from the WebSocket server, and
// submitting the diff produced by function flush to the server.
function newWs(processor, filledOverlayCells) {
  let websocket;
  let scheme;
  if (document.location.protocol === "https:") {
    scheme = "wss";
  } else {
    scheme = "ws"
  }

  function connect() {
    websocket = new WebSocket(`${scheme}://${document.location.host}`);
    websocket.addEventListener("message", processor.enqueue);
  }

  function disconnect(reason) {
    websocket.close(1000, reason);
    processor.clearBuffer();
  }

  function submit() {
    if (filledOverlayCells.size === 0 ||
      websocket === undefined ||
      websocket.readyState !== 1) {
      // There are no changes to submit, or connect hasn't been called yet,
      // or the connection is not in the OPEN state.
      return;
    }
    const diff = flush(filledOverlayCells);
    websocket.send(JSON.stringify(diff));
  }

  return { connect, disconnect, submit };
}

// Processor maintains a FIFO queue of incoming board diffs.
//
// Function enqueue takes a WebSocket message and adds the diff contained
// within to buffer. Diffs are dequeued, parsed, and applied to the board at a
// regular interval.
//
// When the empty diff ("{}") is dequeued, signaling the end of a stream,
// dequeueing stops. Dequeueing starts up again as soon as another message is
// passed to enqueue.
//
// Processor detects buffer overflow. The buffer is considered to be
// overflowing when it has greater than 5 elements.
//
// As a special case, Processor handles the grid message (the first message on
// a connection) by applying it to the board immediately. This is because the
// server sends the grid immediately. Waiting to process it on the next "tick",
// as if it were a diff, would incur a slight delay.
//
// See protocol.md for more information.
function newProcessor() {

  const dequeueInterval = 170;
  let buffer = [];
  // Prefix with _ to avoid clashing with the similarly named function.
  let _isBufferOverflowing = false;
  let dequeueIntervalID;

  function enqueue(message) {
    message.data.text().then((json) => {
      if (dequeueIntervalID === undefined) {
        dequeueIntervalID = setInterval(dequeue, dequeueInterval);
      }
      if (json.startsWith("[")) {
        // Apply the grid message to the board immediately.
        update(json);
        return;
      }
      buffer.push(json);
      checkForBufferOverflow();
    });
  }

  function dequeue() {
    if (buffer.length === 0) {
      return;
    }
    const json = buffer.shift();
    checkForBufferOverflow();
    if (json === "{}" && buffer.length === 0) {
      // We've reached the end of the current stream and there are no further
      // diffs, so we can stop dequeueing. enqueue will start us dequeuing
      // again when appropriate.
      clearInterval(dequeueIntervalID);
      dequeueIntervalID = undefined;
      return;
    }
    update(json);
  }

  function checkForBufferOverflow() {
    _isBufferOverflowing = buffer.length >= 6;
  }

  function isBufferOverflowing() {
    return _isBufferOverflowing;
  }

  function clearBuffer() {
    buffer = [];
  }

  function stopDequeueing() {
    if (dequeueIntervalID !== undefined) {
      clearInterval(dequeueIntervalID);
      dequeueIntervalID = undefined;
    }
  }

  return { enqueue, clearBuffer, isBufferOverflowing, stopDequeueing };
}

// Balancer periodically checks for buffer overflow, and resets the connection
// when this is the case. When Balancer is stopped, it will in turn stop
// Processor.
function newBalancer(ws, processor) {

  const timeBetweenBalances = 8000;
  let balanceBufferTimeoutID;

  function start() {
    balanceBufferTimeoutID = setTimeout(balanceBuffer, timeBetweenBalances);
  }

  function stop() {
    processor.stopDequeueing();
    clearTimeout(balanceBufferTimeoutID);
    balanceBufferTimeoutID = undefined;
  }

  function balanceBuffer() {
    if (processor.isBufferOverflowing()) {
      ws.disconnect("buffer overflow");
      ws.connect();
    }
    balanceBufferTimeoutID = setTimeout(balanceBuffer, timeBetweenBalances);
  }

  return { start, stop };
}

function initBoardCells() {
  const board = document.getElementById("board");
  board.appendChild(makeCells((cell, x, y) => {
    cell.id = `${x},${y}`;
    cell.className = "board_cell_empty";
  }));
}

function initModal(iconButtons) {
  // Allow the modal to be opened and closed.
  const modal = document.getElementById("modal_container");
  iconButtons.namedItem("info").addEventListener("click", () => {
    modal.style.visibility = "";
  });
  iconButtons.namedItem("close").addEventListener("click", () => {
    modal.style.visibility = "hidden";
  });
}

function initView(view) {
  // Prevent the view from scrolling when the mouse is pressed down inside the
  // view and then dragged to the edge of the view.
  let isDraggingView = false;
  view.addEventListener("mousedown", () => {
    isDraggingView = true;
  });
  view.addEventListener("mousemove", (e) => {
    if (isDraggingView) {
      e.preventDefault();
    }
  });
  document.addEventListener("mouseup", () => {
    isDraggingView = false;
  });
}

// initModeSwitch allows switching between pan mode and draw/erase mode.
function initModeSwitch(iconButtons, mouseDraw, touchDraw, tapDraw, mousePan) {
  // isPanMode is true when we are in pan mode, and false when we are in
  // draw/erase mode.
  let isPanMode = false;
  mouseDraw.enable();
  touchDraw.enable();
  tapDraw.enable();

  const move = iconButtons.namedItem("move");
  move.addEventListener("click", () => {
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
}

function initVisChangeHandling(ws, balancer) {
  // Release resources when the page is hidden, and reallocate them when the page
  // becomes visible again.
  // We may get a visibilitychange event with visibilityState "visible" without
  // a corresponding "hidden" event. This was observed to happen on Safari for
  // iOS when minimizing the browser and then quickly maximizing it. So we use
  // isPageHidden to determine whether the visibility state actually changed from
  // "hidden" to "visible", thereby preventing a resource leak.
  let isPageHidden = false;
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "hidden") {
      balancer.stop();
      ws.disconnect("page hidden");
      isPageHidden = true;
    } else if (document.visibilityState === "visible") {
      if (isPageHidden) {
        balancer.start();
        ws.connect();
        isPageHidden = false;
      }
    }
  });
}

// update applies a grid or diff to the board.
function update(json) {
  const change = JSON.parse(json);
  for (const x in change) {
    for (const y in change[x]) {
      const cell = document.getElementById(`${x},${y}`);
      const species = change[x][y];
      if (species !== "") {
        cell.className = "board_cell_filled";
        cell.style.backgroundColor = species;
      } else {
        cell.className = "board_cell_empty";
        cell.style.backgroundColor = "";
      }
    }
  }
}

function makeCells(callback) {
  const frag = document.createDocumentFragment();
  for (let x = 0; x < 120; x++) {
    for (let y = 0; y < 120; y++) {
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
