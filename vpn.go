package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// VPN represents
type VPN struct {
	listenTCPPort int
	uiPort        int

	masterSecret string
}

// NewVPN returns a new uninitialized VPN
func NewVPN(tcpPort, uiPort int) *VPN {
	return &VPN{
		listenTCPPort: tcpPort,
		uiPort:        uiPort,
	}
}

// setMasterSecret sets the master secret on
// passphrase on the server
func (v *VPN) setMasterSecret(secret string) {
	v.masterSecret = secret
}

// start runs the vpn TCP service
func (v *VPN) start() error {
	// establish tcp listener for the vpn service
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", v.listenTCPPort))
	if err != nil {
		return fmt.Errorf("could not establish tcp listener: %s", err)
	}
	log.Printf("[vpn] started tcp listener on :%d", v.listenTCPPort)
	// accept and handle clients
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[vpn] failed to accept tcp connection: %s", err)
			continue
		}
		defer conn.Close()
		log.Println("[vpn] accepted new tcp connection")
		// dispatch new reader and writer
		go v.writer(conn)
		go v.reader(conn)
	}
}

func (v *VPN) writer(conn net.Conn) {
	for {
		msg := "I'm Alice"
		err := writeToConn(conn, msg, v.masterSecret)
		if err != nil {
			log.Printf("[vpn] could not send message to client: %s", err)
			return
		}
		log.Printf("[vpn] sent message: %s", msg)
		time.Sleep(time.Second * 1)
	}
}

func (v *VPN) reader(conn net.Conn) {
	for {
		msg, err := readFromConn(conn, v.masterSecret)
		if err != nil {
			if err == io.EOF {
				log.Printf("[vpn] connection finished (%s) - dropping client", err)
			} else {
				log.Printf("[vpn] could not read from conn: %s", err)
			}
			return
		}
		log.Printf("[vpn] received message: %s", msg)
	}
}
