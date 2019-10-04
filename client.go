package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
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

	// TODO: key exchange

	// return secure client
	return &Client{
		vpnHost: host,
		vpnPort: port,
		conn:    connection,
	}, nil
}

func (c *Client) close() {
	c.conn.Close()
}

func (c *Client) setMasterSecret(s string) {
	c.Lock()
	defer c.Unlock()
	c.masterSecret = s
}

func (c *Client) writer() {
	for {
		if err := c.writeMessage("I'm Bob"); err != nil {
			log.Printf("[client] failed to send message to client: %s", err)
			return
		}
		time.Sleep(time.Second * 10)
	}
}

func (c *Client) reader() {
	for {
		msg, err := readFromConn(c.conn, c.masterSecret)
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Printf("[client] error reading from vpn: %s", err)
			return
		}
		log.Printf("[client] received message: %s", msg)
	}
}

func (c *Client) writeMessage(msg string) error {
	err := writeToConn(c.conn, msg, c.masterSecret)
	if err != nil {
		return err
	}
	log.Printf("[client] sent message: %s", msg)
	return nil
}
