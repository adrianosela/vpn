package main

import (
	"log"
)

var version string // injected at build-time

func main() {
	vpn, err := NewVPN(&Config{
		ListenTCPPort: 80,
		MaxTunnels:    25,
	})
	if err != nil {
		log.Fatalf("could not get new vpn: %s", err)
	}
	if err = vpn.Start(); err != nil {
		log.Fatalf("failed to start vpn: %s", err)
	}
}
