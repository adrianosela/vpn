package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// serveApp serves data based on application state, not on routes
func (a *App) serveApp(w http.ResponseWriter, r *http.Request) {
	switch a.state {
	case stateSetMode:
		a.modeSelectHandler(w, r)
		return
	case stateSetConfig:
		switch a.mode {
		case modeClient:
			a.clientConfigSetHandler(w, r)
			return
		case modeServer:
			a.serverConfigSetHandler(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case stateListenTCP:
		if a.mode != modeServer {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		a.serverListenTCP(w, r)
		return
	case stateDialTCP:
		if a.mode != modeClient {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		a.clientDialTCP(w, r)
		return
	case stateWaitForClient:
		if a.mode != modeServer {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if a.conn == nil {
			a.stateData = fmt.Sprintf("%s<br>$ ...still waiting for client", a.stateData)
			a.serveStateStep(w, r)
			return
		}
		a.stateData = fmt.Sprintf("%s<br>$  client connected! next: generating server ecdh keys", a.stateData)
		a.state = stateGenerateDH
		a.serveStateStep(w, r)
		return
	case stateWaitForServerKey:
		if a.mode != modeClient {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if a.keyExchange.peerPub == nil {
			a.stateData = fmt.Sprintf("%s<br>$ ...still waiting for server key", a.stateData)
			a.serveStateStep(w, r)
			return
		}

		b64peerPub := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.peerPub))
		a.stateData = fmt.Sprintf("%s<br>$ server key received [pub:%s]! next: generating client ecdh keys", a.stateData, b64peerPub)
		a.state = stateGenerateDH
		a.serveStateStep(w, r)
		return
	case stateGenerateDH:
		switch a.mode {
		case modeClient:
			a.clientGenerateDHHandler(w, r)
			return
		case modeServer:
			a.serverGenerateDHHandler(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case stateSendKey:
		switch a.mode {
		case modeClient:
			a.clientSendKeyHandler(w, r)
			return
		case modeServer:
			a.serverSendKeyHandler(w, r)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case stateWaitForClientKey:
		if a.mode != modeServer {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if a.keyExchange.peerPub == nil {
			a.stateData = fmt.Sprintf("%s<br>$ ...still waiting for client key", a.stateData)
			a.serveStateStep(w, r)
			return
		}

		b64peerPub := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.peerPub))
		a.stateData = fmt.Sprintf("%s<br>$ client key received [pub:%s]! next: establishing shared secret", a.stateData, b64peerPub)
		a.state = stateCreateSharedSecret
		a.serveStateStep(w, r)
		return
	case stateCreateSharedSecret:
		a.serveSharedSecret(w, r)
		return
	case stateChat:
		a.serveChat(w, r)
	}
}

func (a *App) modeSelectHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte(modeHTML))
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not parse post form", http.StatusBadRequest)
			return
		}
		if a.mode = r.FormValue("mode"); a.mode != modeClient && a.mode != modeServer {
			http.Error(w, "no mode selected", http.StatusBadRequest)
			return
		}
		a.state = stateSetConfig
		http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *App) serveSharedSecret(w http.ResponseWriter, r *http.Request) {
	if err := a.keyExchange.ComputeSharedSecret(); err != nil {
		log.Fatalf("could not compute shared secret: %s", err)
	}
	b64sharedKey := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.sharedSecret[:]))
	message := fmt.Sprintf("$ generated shared key: %s", b64sharedKey)
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)

	tcpRxChan := make(chan []byte)
	tcpTxChan := make(chan []byte)

	// this thread reads messages from the the TCP connection
	// onto the TCP receive channel and writes messages from the
	// TCP transmission channel to the TCP connection
	// It also writes to the TCP connection from the TCP
	go tcpConnHandler(a.conn, tcpRxChan, tcpTxChan)

	// this thread reads messages from the TCP transmission
	// channel, then b64 decodes and decrypts it, and finally
	// forwards the plaintext to the websocket transmission channel
	// (to then be displayed in the UI by the UI thread)
	go decodeAndDecrypt(tcpRxChan, a.wsTxChan, string(a.keyExchange.sharedSecret[:]))

	// this thread reads messages from the websocket receive
	// channel, then encrypts and b64 encodes it, and finally
	// forwards the b64-encoded-ciphertext to the TCP
	// transmission channel
	go encryptAndEncode(a.wsRxChan, tcpTxChan, string(a.keyExchange.sharedSecret[:]))

	// open chat on next step
	a.state = stateChat
	a.serveStateStep(w, r)
}

func (a *App) serveChat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(chatHTML))
	return
}

func (a *App) serveStateStep(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(messageTemplateHTML, a.stateData)))
	return
}

func (a *App) serveWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	wsConn.WriteMessage(websocket.TextMessage, []byte("Welcome! Enjoy chatting over a secure channel!"))
	go wsConnHandler(wsConn, a.wsRxChan, a.wsTxChan)
}
