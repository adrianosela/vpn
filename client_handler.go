package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func (a *App) clientConfigSetHandler(w http.ResponseWriter, r *http.Request) {

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
		a.state = stateDialTCP
		http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
		return

	default:
		w.WriteHeader(http.StatusNotFound)
		return

	}
}

func (a *App) clientDialTCP(w http.ResponseWriter, r *http.Request) {
	// establish tcp conn
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", a.vpnHost, a.vpnPort), time.Second*10)
	if err != nil {
		log.Fatalf("could not establish tcp connection to vpn server: %s", err)
	}
	log.Printf("[client] established tcp connection with %s:%s", a.vpnHost, a.vpnPort)
	a.conn = conn

	a.stateData = fmt.Sprintf("established tcp connection with %s:%s ...waiting to receive server's public key", a.vpnHost, a.vpnPort)
	a.state = stateWaitForServerKey
	a.serveStateStep(w, r)
}

func (a *App) clientGenerateDHHandler(w http.ResponseWriter, r *http.Request) {
	// generate a fresh ecdh key pair
	if err := a.keyExchange.GenerateKey(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message := fmt.Sprintf("generated diffie hellman keys!")
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)
	a.serveStateStep(w, r)
}
