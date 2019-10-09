package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// App represents application configuration and state
type App struct {
	wsRxChan chan []byte
	wsTxChan chan []byte
	uiPort   int
	conn     net.Conn

	listener net.Listener
	vpnHost  string
	vpnPort  string

	masterSecret string
	mode         string
	state        string
}

const (
	stateSetMode       = "SET MODE"
	stateSetConfig     = "SET CONFIG"
	stateSetPassphrase = "SET PASSPHRASE"
	stateDiffieHellman = "DIFFIE HELLMAN"

	modeServer = "Server"
	modeClient = "Client"
)

func newApp(uiPort int) *App {
	return &App{
		wsRxChan: make(chan []byte),
		wsTxChan: make(chan []byte),
		uiPort:   uiPort,
		state:    stateSetMode,
	}
}

func (a *App) start() {
	// set app HTTP endpoints and websocket handler
	rtr := mux.NewRouter()
	rtr.Path("/app").HandlerFunc(a.serveApp)
	rtr.Methods(http.MethodGet).Path("/chat").HandlerFunc(a.serveChat)
	rtr.Methods(http.MethodGet).Path("/ws").HandlerFunc(a.serveWS)
	// wait a sec to allow server to start, then open browser
	go func() {
		time.Sleep(time.Second * 1)
		if err := openbrowser(fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort)); err != nil {
			log.Fatalf("[error] could not open browser: %s", err)
		}
	}()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", a.uiPort), rtr); err != nil {
		log.Fatal(err)
	}
}

func (a *App) close() {
	if a.conn != nil {
		a.conn.Close()
	}
	if a.listener != nil {
		a.listener.Close()
	}
}

// serveApp serves data based on application state, not on routes
func (a *App) serveApp(w http.ResponseWriter, r *http.Request) {
	switch a.state {
	case stateSetMode:
		a.modeSelectHandler(w, r)
		return
	case stateSetConfig:
		a.configSetHandler(w, r)
		return
	case stateDiffieHellman:
		a.diffieHellmanHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *App) serveChat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(chatHTML))
	return
}

func (a *App) serveWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	go wsConnHandler(wsConn, a.wsRxChan, a.wsTxChan)
}
