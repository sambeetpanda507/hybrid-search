package main

import (
	"fmt"

	"github.com/sambeetpanda507/advance-search/broker-service/internal/server"
)

func main() {
	fmt.Println("Broker Service")
	server.StartServer()
}
