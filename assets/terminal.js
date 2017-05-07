function init() {
  const socket = new WebSocket('ws://localhost:6060/ws');
  socket.addEventListener('message', function (event) {
    var data = JSON.parse(event.data);
    console.log("" + data.y + ": " + data.html);
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

init();

