package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/network/protocol"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/ui/terminal"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatalf("Failed to connect to server: %v", err)
    }
    defer conn.Close()

    display := terminal.NewDisplay()

    fmt.Println("Connected to PokeWorld!")
    for {
        // Simulate receiving world state updates
        var worldState protocol.WorldStatePayload
        if err := protocol.ReadPayload(conn, &worldState); err != nil {
            fmt.Printf("Error reading world state: %v\n", err)
            os.Exit(1)
        }

        // Render the world state
        display.RenderWorldState(&worldState)

        // Accept commands from the player
        fmt.Print("\nEnter command (e.g., move up, catch 001, battle opponent_id): ")
        var commandType string
        var argument string
        fmt.Scanln(&commandType, &argument)

        switch commandType {
        case "move":
            payload := map[string]string{"direction": argument}
            if err := protocol.SendCommand(conn, protocol.CmdMove, payload); err != nil {
                fmt.Printf("Error sending move command: %v\n", err)
                os.Exit(1)
            }
            
        case "catch":
            payload := map[string]string{"pokemon_number": argument}
            if err := protocol.SendCommand(conn, protocol.CmdCatch, payload); err != nil {
                fmt.Printf("Error sending catch command: %v\n", err)
                os.Exit(1)
            }            
        case "battle":
            payload := map[string]string{"opponent_id": argument}
            if err := protocol.SendCommand(conn, protocol.CmdBattle, payload); err != nil {
                fmt.Printf("Error sending battle command: %v\n", err)
                os.Exit(1)
            }
        case "heartbeat":
            if err := protocol.SendCommand(conn, protocol.CmdHeartbeat, nil); err != nil {
                fmt.Printf("Error sending heartbeat command: %v\n", err)
                os.Exit(1)
            }
        default:
            fmt.Println("Unknown command type")
        }
    }
}
