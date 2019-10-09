package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

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

func (a *App) configSetHandler(w http.ResponseWriter, r *http.Request) {
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

			// establish tcp listener for the vpn service
			go func() {
				ln, err := net.Listen("tcp", fmt.Sprintf(":%s", a.vpnPort))
				if err != nil {
					log.Fatal(fmt.Errorf("could not establish tcp listener: %s", err))
				}
				a.listener = ln
				log.Printf("[vpn] started tcp listener on :%s", a.vpnPort)
			}()

			a.state = stateDiffieHellman
			http.Redirect(w, r, fmt.Sprintf("%s:%d/app", "http://localhost", a.uiPort), http.StatusSeeOther)
			return

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

	}
}

func (a *App) diffieHellmanHandler(w http.ResponseWriter, r *http.Request) {
	if a.mode == modeClient {
		// TODO
	} else {
		// TODO
	}
}
