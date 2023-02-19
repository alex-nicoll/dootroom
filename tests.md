Clients to test:
- Asus c302ca, ChromeOS
- iPhone 12, iOS 15, Safari

- Minimize the browser for 5 seconds so that the page is completely hidden, then bring it back into view. The board should still be updating. This can be observed by drawing and submitting changes. (previous bug: board would stop updating due to the WebSocket connection being closed)
- User can press the info button to open the modal, and press the close button to close the modal
- User cannot interact with the view while the modal is displayed. (The reason is that the modal is centered using a invisible container element that takes up the full width of the page. The container element should then take up the full height of the view, because it would be weird for the user to be able to interact with only part of the view while the modal is displayed.)
- If the modal content does not fit inside the modal, then the modal is scrollable. The close button remains visible as the modal is scrolled (previous bug).
- User can press the species button to open a color picker, then select a new color from the color picker, then draw using the new color.
- When the mouse is positioned over an icon button, the icon button changes state. When the mouse is moved off of the icon button, the icon button reverts back to its original state.
- When a finger is placed on an icon button, the icon button changes state. When the finger is removed, the icon button reverts back to it its original state (previous bug).
- The page is not scrollable. Header, view, and controls are completely visible.
- When the submit button is pressed without having drawn any live cells, the game continues to update and the user can continue to submit changes (previous bug: the WebSocket connection would close due to an invalid diff being sent to the server; the game would stop updating and the user would be unable to submit changes).

Draw/erase mode:
- User can tap on cells with one finger to draw and erase.
- User can place one finger on the screen and move it to draw and erase.
- User can click on cells to draw and erase.
- User can drag with the mouse to draw and erase.
- User can position the mouse over the overlay and then use two fingers on a trackpad to pan.
- Tapping with two fingers does not draw or erase.
- Moving two or more fingers across the overlay invokes the browser's default behavior: panning. No cells are drawn.
- Pinching with two or more fingers over the overlay invokes the browser's default behavior: zooming. No cells are drawn.
- Placing one finger on an empty overlay cell and then swiping a second finger outside does not cause any UI elements to change state (previous bug: the swiped element would become filled in).
- When the user drags with a mouse inside a single cell, the cell changes state. When the user then releases the mouse inside that same cell, the cell does not change state (previous bug).
- When the user drags with a mouse to draw or erase several cells and returns the mouse to the starting cell and releases, the starting cell should remain drawn/erased not revert back to its previous state (previous bug).
- Clicking on the overlay padding does not cause the entire overlay to change state (previous bug).
- Pressing the mouse down on an empty overlay cell and then dragging the mouse over the overlay padding does not cause the entire overlay to change state. Further dragging the mouse over a non-overlay element does not cause that element to change state.
- Tapping on the overlay padding with one finger does not cause the entire overlay to change state.
- Touching one finger to an empty overlay cell and then sliding the finger over the overlay padding does not cause the entire overlay to change state. Further sliding the finger over a non-overlay element does not cause that element to change state. Further sliding the finger outside of the viewport does not cause an error to be logged (previous bug).
- Tapping the move button causes it to gain a border, and does not cause any of the other buttons in the control area to move (previous bug).
- Draw something, switch tabs/windows, switch back and immediately submit changes before the board starts to update. The changes should remain on the overlay, unsubmitted (previous bug: the changes disappear and an error is logged to the console).

Pan mode:
- User can place one finger on the overlay and move it to pan.
- User can drag with the mouse to pan.
- Dragging the mouse to the edge of the view does not cause the view to scroll (previous bug).
