var terminalElement;
var opened = false;
var socket;

function attemptReconnection(attempt) {
  if (opened) {
    return;
  }
  if (socket) {
    socket.close();
  }
  console.log("attempting reconnection number " + attempt);
  connect();
  setTimeout(attemptReconnection, 2000, attempt + 1)
}

function connect() {
  terminalElement = document.getElementById('terminal');
  wsURL = new URL(document.URL);
  wsURL.protocol = 'ws:'
  wsURL.pathname = "ws"
  socket = new WebSocket(wsURL.href);
  console.log("Connecting ..")
  socket.addEventListener('message', function (event) {
    var data = JSON.parse(event.data);
    if(data.clientCount > 0) {
       document.getElementById('connected').innerHTML = "" + data.clientCount;
       return;
    }
    var elemId = "line" + data.y
    var elem = document.getElementById(elemId);
    if (!elem) {
      elem = document.createElement('div');
      elem.id = elemId;
      terminalElement.appendChild(elem);
      elem.scrollIntoView(false);
    }
    elem.innerHTML = data.html;
  });
  socket.addEventListener('open', function (event) {
    console.log("Connection open");
    terminalElement.innerHTML = "";
    opened = true;
  });
  socket.addEventListener('close', function (event) {
    console.log("Connection closed:" + event.code);

    if (opened) {
      opened = false;
      setTimeout(attemptReconnection, 2000, 1)
    }
  });

}

