package main

import (
	"bufio"
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
	a.stateData = fmt.Sprintf("$ started tcp listener on :%s, waiting for client...", a.vpnPort)
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
	message := fmt.Sprintf("$ generated diffie hellman keys: [priv:%s] [pub:%s]", b64priv, b64pub)
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)

	a.state = stateWaitForClient
	if a.conn != nil {
		a.state = stateSendKey
	}
	a.serveStateStep(w, r)
}

func (a *App) serverSendKeyHandler(w http.ResponseWriter, r *http.Request) {

	encryptedPub, err := aesEncrypt(a.keyExchange.pub[:], a.masterSecret)
	if err != nil {
		log.Fatalf("could not encrypt server pub key: %s", err)
		return
	}
	b64pub := []byte(b64.StdEncoding.EncodeToString(encryptedPub))

	if _, err := a.conn.Write(append(b64pub, '\n')); err != nil {
		log.Fatalf("could not write key through tcp conn: %s", err)
		return
	}
	message := fmt.Sprintf("$ sent public key to client: [pub:%s] ... waiting for client's public key", b64pub)
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)

	go func() {
		bufReader := bufio.NewReader(a.conn)
		for {
			// read until newline
			encodedPeerPub, err := bufReader.ReadBytes('\n')
			if err != nil {
				log.Fatalf("could not read conn to newline: %s", err)
			}
			// base64 decode the ciphertext (chop off newline)
			decodedPeerPub, err := b64.StdEncoding.DecodeString(string([]byte(encodedPeerPub)[:len(encodedPeerPub)-1]))
			if err != nil {
				log.Fatalf("could not b64 decode message: %s", err)
			}
			// AES decrypt
			decryptedPub, err := aesDecrypt(decodedPeerPub, a.masterSecret)
			if err != nil {
				log.Fatalf("could not decrypt server pub key: %s", err)
				return
			}

			a.keyExchange.peerPub = decryptedPub
			return
		}
	}()

	a.state = stateWaitForClientKey
	a.serveStateStep(w, r)
}
