package server

import (
	"fmt"
	"net/http"
)

func (app *Config) routes() http.Handler {
	mux := http.NewServeMux()

	// GET /ping
	mux.Handle("GET /ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "PONG")
	}))

	// GET /read-csv
	mux.Handle("GET /read-csv", http.HandlerFunc(app.handleReadCSV))

	// GET /store-csv
	mux.Handle("GET /store-csv", http.HandlerFunc(app.handleStoreCSV))

	// POST /get-vector
	mux.Handle("POST /get-vector", http.HandlerFunc(app.handleGetVector))

	// POST /symantic-search
	mux.Handle("POST /symantic-search", http.HandlerFunc(app.handleSymanticSearch))

	// POST /lexical-search
	mux.Handle("POST /lexical-search", http.HandlerFunc(app.handleLexicalSearch))

	// POST /hybrid-search
	mux.Handle("POST /hybrid-search", http.HandlerFunc(app.handleHybridSearch))
	return mux
}
