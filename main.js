let ws = new WebSocket(`ws:\/\/${document.location.host}`);
let elementDoots = document.getElementById("doots");
ws.onmessage = (msg) => {
  let doots = parseInt(elementDoots.innerText, 10);
  elementDoots.innerText = doots + 1;
  msg.data.text().then((text) => console.log(JSON.parse(text)))
}
let elementBtn = document.getElementById("doot-button");
elementBtn.addEventListener("click", (event) => {
  ws.send("doot");
});
window.ws = ws;
