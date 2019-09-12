package main

import (
	"log"
	"net/http"

	"github.com/adrianosela/vpn/service"
)

var version string // injected at build-time

func main() {
	conf := &service.Configuration{}
	router := service.GetRouter(conf)

	log.Println("[INFO] Listening on http://localhost:80")
	if err := http.ListenAndServe(":80", router); err != nil {
		log.Fatalf("[ERROR] ListenAndServe error: %s", err)
	}
}
