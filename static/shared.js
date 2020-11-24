onconnect = function(e) {
  var port = e.ports[0];
  var remembered = {}

  port.onmessage = function(e) {
    if (e.data !== {}) {
      remembered = e.data
    }
    port.postMessage(remembered);
  }
}
