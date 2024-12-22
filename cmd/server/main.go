package main

import (
	"fmt"
	"log"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/network/tcp"
)

func main() {
	// Initialize the game server
	server, err := tcp.NewGameServer("8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	// Start the server
	fmt.Println("Server started on port 8080")
	if err := server.Start(); err != nil {
		log.Fatalf("Error starting game server: %v", err)
	}
}
