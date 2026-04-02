package server

import (
	"fmt"
	"net/http"
)

func (config *Config) routes() http.Handler {
	mux := http.NewServeMux()

	// GET /ping
	mux.Handle("GET /ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "PONG")
	}))

	// POST /synatic-search
	mux.Handle("POST /broker", http.HandlerFunc(config.handleBroker))

	return mux
}
