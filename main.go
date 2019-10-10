package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	// injected at build-time
	version string
	// runtime flag
	uiport = flag.Int("uiport", 8080, "tcp port for application's UI")
)

// App represents application configuration and state
type App struct {
	wsRxChan chan []byte
	wsTxChan chan []byte

	uiPort  int
	conn    net.Conn
	vpnPort string

	masterSecret string
	keyExchange  DH

	mode      string
	state     string
	stateData string

	// client mode only
	vpnHost string
}

const (
	modeServer = "server"
	modeClient = "client"
)

func newApp(uiPort int) *App {
	return &App{
		wsRxChan:    make(chan []byte),
		wsTxChan:    make(chan []byte),
		keyExchange: DH{},
		uiPort:      uiPort,
		state:       stateSetMode,
	}
}

func (a *App) start() {
	// set app HTTP endpoints and websocket handler
	rtr := mux.NewRouter()
	rtr.Path("/app").HandlerFunc(a.serveApp)
	rtr.Methods(http.MethodGet).Path("/chat").HandlerFunc(a.serveChat)
	rtr.Methods(http.MethodGet).Path("/ws").HandlerFunc(a.serveWS)

	go func() {
		time.Sleep(time.Millisecond * 25) // wait to allow server to start, then open browser
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
}

func main() {
	flag.Parse()

	app := newApp(*uiport)
	defer app.close()

	app.start()
}
