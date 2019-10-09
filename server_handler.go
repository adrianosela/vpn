package main

import (
	b64 "encoding/base64"
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
		log.Printf("[server] started tcp listener on :%s", a.vpnPort)
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalf("could not accept tcp conn: %s", err)
			}
			log.Println("[server] accepted client tcp connection")
			a.conn = conn
			break
		}
	}()

	// tell user about it
	a.state = stateWaitForClient
	a.stateData = fmt.Sprintf("started tcp listener on :%s, waiting for client...", a.vpnPort)
	a.serveStateStep(w, r)
}

func (a *App) serverGenerateDHHandler(w http.ResponseWriter, r *http.Request) {
	// generate a fresh ecdh key pair
	if err := a.keyExchange.GenerateKey(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b64priv := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.priv[:]))
	b64pub := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.pub[:]))
	message := fmt.Sprintf("<br>generated diffie hellman keys:<br>[priv:%s] [pub:%s]", b64priv, b64pub)
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)

	a.state = stateWaitForClient
	if a.conn != nil {
		a.state = stateExchangeKeys
	}
	a.serveStateStep(w, r)
}
