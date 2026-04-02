package main

import (
	"fmt"
	"log"

	"github.com/sambeetpanda507/advance-search/search-service/internal/database"
	"github.com/sambeetpanda507/advance-search/search-service/internal/server"
)

func main() {
	fmt.Println("Search Service")

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Panic(err)
		return
	}

	// Start the server
	server.StartServer(db)
}
