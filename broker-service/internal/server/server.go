package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
}

func StartServer() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Panic("Port is required")
	}

	// Initialize the app config
	appConfig := &Config{}

	// Initialize the server
	svr := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: enableCors(logger(appConfig.routes())),
	}

	quit := make(chan os.Signal, 1)
	go func() {
		log.Println("Starting server")
		if err := svr.ListenAndServe(); err != nil {
			log.Panic(err)
		}

		log.Printf("Server is listening on port: %s\n", port)
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
	log.Println("Bye")
}
