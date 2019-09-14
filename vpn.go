package main

import (
	"bytes"
	"crypto/rsa"
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
	// receive client encryption key
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	clientKey, err := decodePubKeyPEM(buf.Bytes())
	if err != nil {
		log.Printf("could not decode client key - kicking client: %s", err)
		return
	}
	log.Printf("[vpn] received client key %s", buf.Bytes())
	// add new key to keystore and schedule removal on return
	kid := getKeyID(clientKey)
	v.kstore[kid] = clientKey
	defer func() { delete(v.kstore, kid) }()
	// send server public key
	if _, err := conn.Write(encodePubKeyPEM(&v.k.PublicKey)); err != nil {
		log.Printf("could not send server key - kicking client: %s", err)
		return
	}
	log.Println("[vpn] sent server key to new client")
	time.Sleep(time.Second * 5)
	// receive encrypted KEY ACK message
	var buf2 bytes.Buffer
	io.Copy(&buf2, conn)
	plain, err := decryptMessage(buf2.Bytes(), v.k)
	if err != nil {
		log.Printf("could not decrypt expected KEY ACK message - kicking client: %s", err)
		return
	}
	if string(plain) != "KEY ACK" {
		log.Printf("expected KEY ACK message, got %s", string(plain))
		return
	}
	for {
		fmt.Println("SECURE LINK ESTABLISHED")
		time.Sleep(time.Minute * 1)
	}
}
