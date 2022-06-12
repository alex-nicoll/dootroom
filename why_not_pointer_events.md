# Why I'm not using pointer events (yet)

- Pointer events aren't supported in older browers, like Safari 12 from 2019.
- We only need to support drawing with a single touch. e.touches.length === 1
  is more intuitive than e.isPrimary. Not to mention that these checks
  shouldn't be needed when the pointer is a mouse...
- I couldn't get "releasing of implicit pointer capture" to work, so it would
  appear that touch-specific manual hit testing is still needed. In retrospect,
  I might have needed to add the gotpointercapture event handler to, and call
  releasePointerCapture on, the individual cell element that captured the
  pointer. Which would mean adding an event handler to every cell, which would
  require some refactoring.
- e.preventDefault() in the pointermove event handler does not prevent the
  browser from scrolling when we draw. So we would still need a touchmove event
  handler. Or, we could take the recommended approach of setting
  touch-action:none in CSS. This would unfortunately disable multi-touch pan and
  zoom in draw mode, which are nice features to have; I occasionally use zoom in
  testing. Oh wait, we could actually set touch-action:pinch-zoom.

Still, I think that all of the above, along with the overall obscurity of the
pointer events API, cause me to lean away from it for this particular use case.
I think the code is a bit easier to grasp as is.

One last word on the pointer events API. I've seen several examples in the MDN
docs and the W3C spec of switching on the device type to invoke device-specific
handling. Doesn't this somewhat defeat the purpose of pointer events? Granted,
the spec also mentions that pointer events make it easier to tell whether a
mouse event came from a mouse or another device type. One could invoke some
behavior IFF an actual mouse click occured, as opposed to a simulated mouse
click caused by a tap. This would be a good use case for pointer events.
