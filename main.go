package main

import (
	"flag"
	"log"
)

var (
	// injected at build-time
	version string
	// flag arguments
	clientMode = flag.Bool("c", false, "run application in client mode")
	host       = flag.String("host", "localhost", "vpn host to use")
	port       = flag.Int("port", 80, "tcp port for vpn")
	tunnels    = flag.Int("tunnels", 25, "maximum simultaneous vpn clients")
)

func main() {
	flag.Parse()

	if *clientMode {
		clientMain()
	} else {
		serverMain()
	}
}

func serverMain() {
	vpn, err := NewVPN(&Config{
		ListenTCPPort: *port,
		MaxTunnels:    *tunnels,
		SharedSecret:  mockPassphrase,
	})
	if err != nil {
		log.Fatalf("could not get new vpn: %s", err)
	}
	if err = vpn.Start(); err != nil {
		log.Fatalf("failed to start vpn: %s", err)
	}
}

func clientMain() {
	client := NewClient(*host, *port)
	client.setMasterSecret(mockPassphrase)
	client.start()
}
