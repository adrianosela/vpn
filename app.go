package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
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
	rtr.Methods(http.MethodGet).Path("/secure").HandlerFunc(serveChatHTML)
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
	a.conn.Close()
}

func (a *App) serveModeSelect(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // GET
		w.Write([]byte(modeHTML))
		return
	case http.MethodPost: // POST
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
	default: // reject any other methods
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func serveChatHTML(w http.ResponseWriter, r *http.Request) { w.Write([]byte(chatHTML)) }

func (a *App) serveConfigSet(w http.ResponseWriter, r *http.Request) {
	if a.mode == modeClient {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(clientConfigHTML))
			return
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "could not parse post form", http.StatusBadRequest)
				return
			}
			a.masterSecret = r.FormValue("passphrase")
			a.vpnHost = r.FormValue("host")
			a.vpnPort = r.FormValue("port")
			if a.masterSecret == "" {
				http.Error(w, "passphrase was empty", http.StatusBadRequest)
				return
			}
			if a.vpnHost == "" {
				http.Error(w, "host was empty", http.StatusBadRequest)
				return
			}
			if a.vpnHost == "" {
				http.Error(w, "port was empty", http.StatusBadRequest)
				return
			}
			// establish tcp conn
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", a.vpnHost, a.vpnPort), time.Second*10)
			if err != nil {
				log.Fatalf("could not establish tcp connection to vpn server: %s", err)
			}
			a.conn = conn
			log.Printf("[client] established tcp connection with %s:%s", a.vpnHost, a.vpnPort)

			a.state = stateDiffieHellman
			http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(serverConfigHTML))
			return
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "could not parse post form", http.StatusBadRequest)
				return
			}
			a.masterSecret = r.FormValue("passphrase")
			a.vpnPort = r.FormValue("port")
			if a.masterSecret == "" {
				http.Error(w, "passphrase was empty", http.StatusBadRequest)
				return
			}
			if a.vpnPort == "" {
				http.Error(w, "port was empty", http.StatusBadRequest)
				return
			}
			// FIXME: dispatch serve TCP here
			a.state = stateDiffieHellman
			http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func (a *App) serveApp(w http.ResponseWriter, r *http.Request) {
	switch a.state {
	case stateSetMode:
		a.serveModeSelect(w, r)
		return
	case stateSetConfig:
		a.serveConfigSet(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *App) serveWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	go wsConnHandler(wsConn, a.wsRxChan, a.wsTxChan)
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
