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

// Client of the VPN service
type Client struct {
	vpnHost string
	vpnPort int

	vpnKey    *rsa.PublicKey
	clientKey *rsa.PrivateKey

	conn net.Conn
}

// NewClient establishes a secure connection to the VPN at host:port
func NewClient(host string, port int) (*Client, error) {
	// establish tcp conn
	connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("could not establish tcp connection to vpn server: %s", err)
	}
	log.Printf("[client] established tcp connection with %s:%d", host, port)
	// generate client key
	k, err := generateRSAKey(4096)
	if err != nil {
		return nil, fmt.Errorf("could not generate client key: %s", err)
	}
	// format and transport client key (pub part of client decryption key)
	pubBytes := encodePubKeyPEM(&k.PublicKey)
	if _, err := connection.Write(pubBytes); err != nil {
		return nil, fmt.Errorf("could not send client key to vpn: %s", err)
	}
	log.Println("[client] sent client key to server")
	time.Sleep(time.Second * 5)
	// read vpn key (pub encryption key)
	var buf bytes.Buffer
	io.Copy(&buf, connection)
	vk, err := decodePubKeyPEM(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("could not receive vpn key over tcp conn: %s", err)
	}
	log.Println("[client] vpn server key received")
	// encrypt and send KEY ACK message to server to complete handshake
	msg, err := encryptMessage([]byte("KEY ACK"), vk)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt KEY ACK message: %s", err)
	}
	if _, err = connection.Write(msg); err != nil {
		return nil, fmt.Errorf("could not send KEY ACK message: %s", err)
	}
	// return secure client
	return &Client{
		vpnHost:   host,
		vpnPort:   port,
		vpnKey:    vk,
		clientKey: k,
		conn:      connection,
	}, nil
}

func (c *Client) writeMessage(msg string) error {
	// todo
	return nil
}
