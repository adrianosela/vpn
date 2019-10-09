package main

import (
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
		a.stateData = fmt.Sprintf("%s<br>$ server key received! next: generating client ecdh keys", a.stateData)
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

		a.stateData = fmt.Sprintf("%s<br>$ client key received! next: establishing shared secret", a.stateData)
		a.state = stateCreateSharedSecret
		a.serveStateStep(w, r)
		return
	case stateCreateSharedSecret:
		a.serveChat(w, r) // FIXME
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
