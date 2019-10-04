package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Client of the VPN service
type Client struct {
	sync.Mutex // inherit lock behavior
	vpnHost    string
	vpnPort    int

	masterSecret string

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
	// TODO

	// return secure client
	return &Client{
		vpnHost: host,
		vpnPort: port,
		conn:    connection,
	}, nil
}

func (c *Client) SetMasterSecret(s string) {
	c.Lock()
	defer c.Unlock()
	c.masterSecret = s
}

func (c *Client) writeMessage(msg string) error {
	// todo
	return nil
}
