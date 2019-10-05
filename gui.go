package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type uiConfig struct {
	uiPort int

	wsRxChan chan []byte
	wsTxChan chan []byte
}

type uiData struct {
	Data string `json:"data"` // message body
}

func ui(c *uiConfig) {
	go func() {
		time.Sleep(time.Second * 1)
		if err := openbrowser(fmt.Sprintf("%s:%d/home", "http://localhost", c.uiPort)); err != nil {
			log.Fatalf("[gui] could not open browser for GUI: %s", err)
		}
	}()
	rtr := mux.NewRouter()

	// landing page (ask for passphrase)
	rtr.Methods(http.MethodGet).Path("/home").HandlerFunc(serveHomeHTML)
	rtr.Methods(http.MethodPost).Path("/home").HandlerFunc(c.passphraseHandler)

	// secure chat endpoint
	rtr.Methods(http.MethodGet).Path("/secure").HandlerFunc(serveChatHTML)

	// upgrade to websocket endpoint
	rtr.Methods(http.MethodGet).Path("/ws").HandlerFunc(c.serveWS)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.uiPort), rtr); err != nil {
		log.Fatal(err)
	}
}

// serveHomeHTML serves the home page (where the secret is requested)
func serveHomeHTML(w http.ResponseWriter, r *http.Request) { w.Write([]byte(homeHTML)) }

// passPhraseHandler receives the password from the user
func (c *uiConfig) passphraseHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "could not parse POST form: %s", err)
		return
	}
	// set passphrase globally
	passphrase = r.FormValue("passphrase")
	http.Redirect(w, r, fmt.Sprintf("%s:%d/secure", "http://localhost", c.uiPort), http.StatusSeeOther)
}

// serveChatHTML serves the chat page (where a secure channel has been established)
func serveChatHTML(w http.ResponseWriter, r *http.Request) { w.Write([]byte(chatHTML)) }

// serveWS upgrades HTTP to websockets
func (c *uiConfig) serveWS(w http.ResponseWriter, r *http.Request) {
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// prompt for passphrase
	prompt := []byte(fmt.Sprintf("Welcome! your passphrase was: %s", passphrase))
	if err := wsConn.WriteMessage(websocket.TextMessage, prompt); err != nil {
		log.Fatal(err)
	}

	go wsConnHandler(wsConn, c.wsRxChan, c.wsTxChan)
}

func openbrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return errors.New("unsupported platform")
	}
}

const homeHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>VPN - Secure Channel Service</title>
</head>
<body>
<div id="log"></div>
        <form action="/home" method="post">
            Enter Passphrase:
            <input type="text" name="passphrase">
            <input type="submit" value="Enter">
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
