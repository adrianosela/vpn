package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Client of the VPN service
type Client struct {
	vpnHost string
	vpnPort int

	uiPort int

	masterSecret string

	conn net.Conn
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
	defer conn.Close()
	c.conn = conn
	log.Printf("[client] established tcp connection with %s:%d", c.vpnHost, c.vpnPort)

	// dispatch tcp conn reader and writer threads
	go c.reader()
	go c.writer()

	// dispatch UI thread, wait a sec, open browser
	go c.ui()
	time.Sleep(time.Second * 1)
	if err = openbrowser(fmt.Sprintf("%s:%d/", "http://localhost", c.uiPort)); err != nil {
		return fmt.Errorf("[client] could not open browser for GUI: %s", err)
	}

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
			log.Printf("[client] could not send message to vpn: %s", err)
			return
		}
		log.Printf("[client] sent message: %s", msg)
		time.Sleep(time.Second * 1)
	}
}

func (c *Client) reader() {
	for {
		msg, err := readFromConn(c.conn, c.masterSecret)
		if err != nil {
			if err == io.EOF {
				log.Printf("[client] connection terminated by server (%s) - dropping off", err)
			} else {
				log.Printf("[client] could not read from conn: %s", err)
			}
			return
		}
		log.Printf("[client] received message: %s", msg)
	}
}

func (c *Client) ui() {
	rtr := mux.NewRouter()
	rtr.Methods(http.MethodGet).Path("/").HandlerFunc(serveHTML)
	rtr.Methods(http.MethodGet).Path("/ws").HandlerFunc(serveWS)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.uiPort), rtr); err != nil {
		log.Fatal(err)
	}
}
