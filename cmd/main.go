package main

import (
	"log"

	"github.com/ra-shree/prequal-server/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Printf("Could not start the server: %v", err)
	}
}
