package main

const modeHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>VPN - Secure Channel Service</title>
	</head>
	<body>
		<h2> Select Mode of Operation: </h2> </br>
		<form action="/app" method="post">
			Server:<input type="radio" name="mode" value="server"><br>
			Client:<input type="radio" name="mode" value="client"><br>
  		<input type="submit" value="Enter">
		</form>
	</body>
</html>
`

const clientConfigHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>VPN - Secure Channel Service</title>
	</head>
	<body>
		<h2> Configure Client: </h2><br>
  	<form action="/app" method="post">
    	Server Host: <input type="text" name="host"><br>
			Server Port: <input type="text" name="port"><br>
      Shared Secret Value: <input type="text" name="passphrase">
      <input type="submit" value="Enter">
    </form>
	</body>
</html>
`

const serverConfigHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>VPN - Secure Channel Service</title>
	</head>
	<body>
		<h2> Configure Server: </h2> <br>
		<form action="/app" method="post">
			TCP Listener Port: <input type="text" name="port"><br>
			Shared Secret Value: <input type="text" name="passphrase">
			<input type="submit" value="Enter">
		</form>
	</body>
</html>
`

const messageTemplateHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>VPN - Secure Channel Service</title>
	</head>
	<body>
		<h4> Messages: </h4> %s <br>
		<form action="/app" method="post">
			<input type="submit" value="Continue">
		</form>
	</body>
</html>
`

const chatHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>VPN - Secure Channel Service</title>
<script type="text/javascript">
window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");
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
        conn.send(JSON.stringify({ data: msg.value }));
        msg.value = "";
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
</form>
</body>
</html>
`
