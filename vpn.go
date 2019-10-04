package main

import (
	"fmt"
	"log"
	"net"
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
			log.Printf("failed to accept tcp connection: %s", err)
			continue
		}
		log.Println("[vpn] accepted new tcp connection")
		go v.handleConn(conn)
	}
}

// handleConn establishes a secure communication channel
// and proceeds to handle encrypted requests
func (v *VPN) handleConn(conn net.Conn) {
	defer conn.Close()
	// TODO
}
