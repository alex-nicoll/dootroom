Clients to test:
- Asus c302ca, ChromeOS
- iPhone 12, iOS 15, Safari

- User can tap on cells with one finger to draw and erase.
- User can place one finger on the screen and move it to draw and erase.
- User can click on cells to draw and erase.
- User can drag with the mouse to draw and erase.
- Tapping with two fingers does not draw or erase.
- Moving two or more fingers across the screen invokes the browser's default behavior: scrolling.
- Pinching with two or more fingers invokes the browser's default behavior: zooming. 
- Placing one finger inside the overlay and then swiping a second finger outside does not cause any UI elements to change state (previous bug: the swiped element would become filled in).
- When the user drags with a mouse inside a single cell, the cell should change state. When the user then releases the mouse inside that same cell, the cell does not change state. (previous bug)
- When the user drags with a mouse to draw or erase several cells and returns the mouse to the starting cell and releases, the starting cell should remain drawn/erased not revert back to its previous state. (previous bug)
- Drawing on the overlay element with one finger and then moving that finger off of the overlay does not cause non-cell elements to change state. (previous bug)
- Same as above, but further moving the finger outside of the viewport does not cause an error to be logged. (previous bug)
- Minimize the browser for 5 seconds so that the page is completely hidden, then bring it back into view. The grid should still be updating. This can be observed by drawing and submitting changes. (previous bug: grid would stop updating due to the WebSocket connection being closed)
