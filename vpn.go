package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
)

// VPN represents
type VPN struct {
	// config
	listenTCPPort int
	maxTunnels    int
	// server decryption key
	k *rsa.PrivateKey
	// map of cient id to encryption key
	kstore map[string]*rsa.PublicKey
}

// NewVPN returns a new uninitialized VPN
func NewVPN(c *Config) (*VPN, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %s", err)
	}
	return &VPN{
		listenTCPPort: c.ListenTCPPort,
		maxTunnels:    c.MaxTunnels,
	}, nil
}

// Start runs the vpn TCP service
func (v *VPN) Start() error {
	// new tcp listener
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", v.listenTCPPort))
	if err != nil {
		return fmt.Errorf("could not establish tcp listener: %s", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println()
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	// - receive client public key
	// - send server public key
	// - receive encrypted KACK (allow up to 10 seconds)
	// ---- secure tunnel established ----
}
