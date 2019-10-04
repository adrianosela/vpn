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
	// config
	listenTCPPort int
	maxTunnels    int
	sharedSecret  string
}

// NewVPN returns a new uninitialized VPN
func NewVPN(c *Config) (*VPN, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid vpn configuration: %s", err)
	}
	return &VPN{
		listenTCPPort: c.ListenTCPPort,
		maxTunnels:    c.MaxTunnels,
		sharedSecret:  c.SharedSecret,
	}, nil
}

// Start runs the vpn TCP service
func (v *VPN) Start() error {
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
		log.Println("[vpn] accepted new tcp connection")
		// dispatch new reader and writer
		go v.writer(conn)
		go v.reader(conn)
	}
}

func (v *VPN) writer(conn net.Conn) {
	defer conn.Close()
	for {
		msg := "I'm Alice"
		err := writeToConn(conn, msg, v.sharedSecret)
		if err != nil {
			log.Printf("[vpn] failed to send message to client: %s", err)
		}
		log.Printf("[vpn] sent message: %s", msg)
		time.Sleep(time.Second * 10)
	}
}

func (v *VPN) reader(conn net.Conn) {
	defer conn.Close()
	for {
		msg, err := readFromConn(conn, v.sharedSecret)
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Printf("[vpn] error reading from client: %s", err)
			return
		}
		log.Printf("[vpn] received message: %s", msg)
	}
}
