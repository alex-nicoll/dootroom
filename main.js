let ws = new WebSocket(`ws:\/\/${document.location.host}`);
let elementDoots = document.getElementById("doots");
ws.onmessage = (msg) => {
  let doots = parseInt(elementDoots.innerText, 10);
  elementDoots.innerText = doots + 1;
  console.log(msg.data);
}
let elementBtn = document.getElementById("doot-button");
elementBtn.addEventListener("click", (event) => {
  ws.send("doot");
});
setInterval(() => ws.send("doot"), 100)
