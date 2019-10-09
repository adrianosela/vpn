package main

import "flag"

var (
	// injected at build-time
	version string
	// runtime flag
	uiport = flag.Int("uiport", 8080, "tcp port for UI's http listener")
)

func main() {
	flag.Parse()

	app := newApp(*uiport)
	defer app.close()

	app.start()
}
