function init() {
  wsURL = new URL(document.URL);
  wsURL.protocol = 'ws:'
  wsURL.pathname = "ws"
  const socket = new WebSocket(wsURL.href);
  console.log("Connecting ..")
  socket.addEventListener('message', function (event) {
    var data = JSON.parse(event.data);
    // console.log("" + data.y + ": " + data.html);
    var elemId = "line" + data.y
    var elem = document.getElementById(elemId);
    if (!elem) {
      elem = document.createElement('div');
      elem.id = elemId;
      document.getElementById('terminal').appendChild(elem);
      elem.scrollIntoView(false);
    }
    elem.innerHTML = data.html;
  });
}

