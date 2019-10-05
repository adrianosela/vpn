package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Client of the VPN service
type Client struct {
	vpnHost      string
	vpnPort      int
	uiPort       int
	masterSecret string
	conn         net.Conn
}

// NewClient establishes a secure connection to the VPN at host:port
func NewClient(host string, port, uiPort int) *Client {
	return &Client{
		vpnHost: host,
		vpnPort: port,
		uiPort:  uiPort,
	}
}

func (c *Client) start() error {
	// establish tcp conn
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.vpnHost, c.vpnPort), time.Second*10)
	if err != nil {
		return fmt.Errorf("could not establish tcp connection to vpn server: %s", err)
	}

	// schedule conn close and add to client
	defer conn.Close()
	c.conn = conn
	log.Printf("[client] established tcp connection with %s:%d", c.vpnHost, c.vpnPort)

	// dispatch tcp conn reader and writer threads
	// go c.reader()
	go c.writer()

	// catch shutdown
	signalCatch := make(chan os.Signal, 1)
	signal.Notify(signalCatch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		<-signalCatch
		log.Printf("[client] shutdown signal received, terminating")
		return nil
	}
}

func (c *Client) setMasterSecret(s string) {
	c.masterSecret = s
}
