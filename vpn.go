package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
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

	// map of encryption keys (indexed by hash)
	kstore map[string]*rsa.PublicKey
}

// NewVPN returns a new uninitialized VPN
func NewVPN(c *Config) (*VPN, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid vpn configuration: %s", err)
	}
	key, err := generateRSAKey(4096)
	if err != nil {
		return nil, fmt.Errorf("could not generate server key: %s", err)
	}
	return &VPN{
		listenTCPPort: c.ListenTCPPort,
		maxTunnels:    c.MaxTunnels,
		k:             key,
		kstore:        make(map[string]*rsa.PublicKey),
	}, nil
}

// Start runs the vpn TCP service
func (v *VPN) Start() error {
	// establish tcp listener for the vpn service
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", v.listenTCPPort))
	if err != nil {
		return fmt.Errorf("could not establish tcp listener: %s", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept tcp connection: %s", err)
			continue
		}
		go v.handleConn(conn)
	}
}

// handleConn establishes a secure communication channel
// and proceeds to handle encrypted requests
func (v *VPN) handleConn(conn net.Conn) {
	defer conn.Close()

	var buf bytes.Buffer
	io.Copy(&buf, conn)
	// todo - receive client public key
	// todo - send server public key
	// todo - receive encrypted KACK (allow up to 10 seconds)
	// ---- secure tunnel established ----
}
