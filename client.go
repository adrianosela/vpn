package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Client of the VPN service
type Client struct {
	vpnHost string
	vpnPort int

	masterSecret string

	conn net.Conn
}

// NewClient establishes a secure connection to the VPN at host:port
func NewClient(host string, port int) *Client {
	return &Client{
		vpnHost: host,
		vpnPort: port,
	}
}

func (c *Client) start() error {
	// establish tcp conn
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.vpnHost, c.vpnPort), time.Second*10)
	if err != nil {
		return fmt.Errorf("could not establish tcp connection to vpn server: %s", err)
	}
	defer conn.Close()
	c.conn = conn
	log.Printf("[client] established tcp connection with %s:%d", c.vpnHost, c.vpnPort)

	// dispatch reader and writer
	go c.reader()
	go c.writer()

	// TODO: open browser console here (unless disabled)

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

func (c *Client) writer() {
	for {
		msg := "I'm Bob"
		err := writeToConn(c.conn, msg, c.masterSecret)
		if err != nil {
			log.Printf("[client] failed to send message to client: %s", err)
			return
		}
		log.Printf("[client] sent message: %s", msg)
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
