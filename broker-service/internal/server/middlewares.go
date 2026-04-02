package server

import (
	"log"
	"net/http"
	"time"
)

func enableCors(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

type CustomWriter struct {
	http.ResponseWriter
	statusCode int
}

func (c *CustomWriter) WriteHeader(statusCode int) {
	c.ResponseWriter.WriteHeader(statusCode)
	c.statusCode = statusCode
}

func logger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		endTime := time.Since(startTime)
		method := r.Method
		endPoint := r.URL.Path
		cw := &CustomWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(cw, r)
		log.Printf("%s %d %s %s", method, cw.statusCode, endPoint, endTime)
	}
}
