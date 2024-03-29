package main

import (
	"bufio"
	b64 "encoding/base64"
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
	a.stateData = fmt.Sprintf("$ established tcp connection with %s:%s ...waiting to receive server's public key", a.vpnHost, a.vpnPort)
	a.state = stateWaitForServerKey
	a.serveStateStep(w, r)
}

func (a *App) clientGenerateDHHandler(w http.ResponseWriter, r *http.Request) {
	// generate a fresh ecdh key pair
	if err := a.keyExchange.GenerateKey(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b64priv := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.priv[:]))
	b64pub := []byte(b64.StdEncoding.EncodeToString(a.keyExchange.pub[:]))
	message := fmt.Sprintf("$ generated diffie hellman keys: [priv:%s] [pub:%s]", b64priv, b64pub)
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)
	a.state = stateSendKey
	a.serveStateStep(w, r)
}

func (a *App) clientSendKeyHandler(w http.ResponseWriter, r *http.Request) {
	encryptedPub, err := aesEncrypt(a.keyExchange.pub[:], a.masterSecret)
	if err != nil {
		log.Fatalf("could not encrypt client pub key: %s", err)
		return
	}
	b64pub := []byte(b64.StdEncoding.EncodeToString(encryptedPub[:]))
	if _, err := a.conn.Write(append(b64pub, '\n')); err != nil {
		log.Fatalf("could not write key to tcp conn: %s", err)
		return
	}
	message := fmt.Sprintf("$ sent encrypted public key to server: ...next: establishing shared secret")
	a.stateData = fmt.Sprintf("%s<br>%s", a.stateData, message)
	a.state = stateCreateSharedSecret
	a.serveStateStep(w, r)
}
