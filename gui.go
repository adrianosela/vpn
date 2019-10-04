package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// serveHTML serves the "homepage"
func serveHTML(w http.ResponseWriter, r *http.Request) { w.Write([]byte(indexHTML)) }

// serveWS upgrades HTTP to websockets
func serveWS(w http.ResponseWriter, r *http.Request) {
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// TODO
	time.Sleep(time.Second * 10)
	defer wsConn.Close()
}

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>VPN - Secure Channel Service</title>
<script type="text/javascript">
window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");
    var to = document.getElementById("to");
    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }
    document.getElementById("form").onsubmit = function () {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(JSON.stringify({ data: msg.value, to: to.value }));
        msg.value = "";
	to.value = "";
        return false;
    };
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };
        conn.onmessage = function (event) {
            var messages = event.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                var item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
        };
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>this browser does not support websockets</b>";
        appendLog(item);
    }
};
</script>
<style type="text/css">
html {
    overflow: hidden;
}
body {
    overflow: hidden;
    padding: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    background: gray;
}
#log {
    background: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    position: absolute;
    top: 0.5em;
    left: 0.5em;
    right: 0.5em;
    bottom: 3em;
    overflow: auto;
}
#form {
    padding: 0 0.5em 0 0.5em;
    margin: 0;
    position: absolute;
    bottom: 1em;
    left: 0px;
    width: 100%;
    overflow: hidden;
}
</style>
</head>
<body>
<div id="log"></div>
<form id="form">
	<input type="submit" value="Send" />
   	<input type="text" id="msg" size="64"/>
	<input type="text" id="to" size="38"/>
</form>
</body>
</html>
`
