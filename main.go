package main

import (
	"flag"
)

var (
	// injected at build-time
	version string

	host = flag.String("host", "localhost", "vpn host to use")
	port = flag.Int("port", 80, "tcp port for vpn")

	clientMode = flag.Bool("c", false, "run application in client mode")
	uiport     = flag.Int("uiport", 8080, "tcp port for UI's http listener")
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
	vpn := NewVPN(*port, *uiport)
	vpn.setMasterSecret(mockPassphrase)
	vpn.start()
}

func clientMain() {
	client := NewClient(*host, *port, *uiport)
	client.setMasterSecret(mockPassphrase)
	client.start()
}
