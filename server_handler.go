package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func (a *App) serverConfigSetHandler(w http.ResponseWriter, r *http.Request) {
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
		a.state = stateListenTCP
		http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *App) serverListenTCP(w http.ResponseWriter, r *http.Request) {
	// establish tcp listener for the vpn service
	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%s", a.vpnPort))
		if err != nil {
			log.Fatal(fmt.Errorf("could not establish tcp listener: %s", err))
		}
		a.listener = ln
		log.Printf("[vpn] started tcp listener on :%s", a.vpnPort)
	}()
	// tell user about it
	a.state = stateGenerateDH
	a.stateData = fmt.Sprintf("started tcp listener on :%s", a.vpnPort)
	a.serveStateStep(w, r)
}

func (a *App) serverGenerateDHHandler(w http.ResponseWriter, r *http.Request) {
	// generate a fresh ecdh key pair
	if err := a.keyExchange.GenerateKey(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message := fmt.Sprintf("generated diffie hellman keys!")
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)
	a.serveStateStep(w, r)
}
