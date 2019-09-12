package service

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Configuration represents necessary parameters
// to serve HTTP requests for the VPN
type Configuration struct{}

//GetRouter returns a router given a Configuration
func GetRouter(conf *Configuration) *mux.Router {

	router := mux.NewRouter()

	// clients' encryption keys discovery
	router.Methods("GET").Path("/keys").HandlerFunc(keysHandler)

	return router
}

func keysHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string("mock key"))
	return
}
