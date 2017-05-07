function init() {
  var ContentElement = document.getElementById('terminal');
  const socket = new WebSocket('ws://localhost:6060/ws');
  socket.addEventListener('message', function (event) {
      console.log(event.data);
  });
  ContentElement.innerHTML = "hello";
}

init();
