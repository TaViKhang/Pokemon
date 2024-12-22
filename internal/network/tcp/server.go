package tcp

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/game/pokebat"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/game/pokecat"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/network/protocol"
)

type GameServer struct {
	listener net.Listener
	world    *pokecat.Grid
	battles  map[string]*pokebat.Battle
	clients  map[string]*Client
	sessions *SessionManager
	mu       sync.RWMutex
	done     chan struct{}
}

func NewGameServer(port string) (*GameServer, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	return &GameServer{
		listener: listener,
		world:    pokecat.NewGrid(),
		battles:  make(map[string]*pokebat.Battle),
		clients:  make(map[string]*Client),
		sessions: NewSessionManager(),
		done:     make(chan struct{}),
	}, nil
}

func (s *GameServer) Start() error {
	defer s.Cleanup()

	go s.monitorBattles()

	for {
		select {
		case <-s.done:
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go s.handleNewConnection(conn)
		}
	}
}

func (s *GameServer) handleNewConnection(conn net.Conn) {
	// Authentication và player initialization
	player := models.NewPlayer(generateID())

	// Tạo client mới với LastPing
	client := NewClient(conn, player)

	// Add to world grid
	if err := s.world.AddPlayer(player); err != nil {
		log.Printf("Failed to add player to world: %v", err)
		conn.Close()
		return
	}

	// Register client
	s.mu.Lock()
	s.clients[player.GetID()] = client
	s.mu.Unlock()

	// Start client handlers
	go s.handleClientRead(client)
	go s.handleClientWrite(client)
	go s.startHeartbeat(client)
}

func (s *GameServer) handleClientRead(client *Client) {
    defer func() {
        s.removeClient(client)
        client.Conn.Close()
    }()

    // Send initial world state to the client
    worldState := s.world.GetState() // Assuming there's a method to get the world state
    err := protocol.SendCommand(client.Conn, protocol.CmdWorldState, worldState)
    if err != nil {
        log.Printf("Error sending world state: %v", err)
        return
    }

    // Handle incoming commands (game actions) from the client
    for {
        buf := make([]byte, 4096)
        n, err := client.Conn.Read(buf)
        if err != nil {
            return
        }

        if err := s.handleCommand(client, buf[:n]); err != nil {
            // Broadcast error to client
            s.broadcast(client.Player.GetID(), "error", err.Error())
            log.Printf("Error handling command: %v", err)
        }
    }
}



func (s *GameServer) handleClientWrite(client *Client) {
	for msg := range client.SendChan {
		if _, err := client.Conn.Write(msg); err != nil {
			log.Printf("Error writing to client: %v", err)
			return
		}
	}
}

func (s *GameServer) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.world.RemovePlayer(client.Player); err != nil {
		log.Printf("Error removing player from world: %v", err)
	}

	delete(s.clients, client.Player.GetID())
	close(client.SendChan)
}

func (s *GameServer) monitorBattles() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			for id, battle := range s.battles {
				if time.Since(battle.StartTime) > time.Duration(constants.BattleTimeout)*time.Second {
					delete(s.battles, id)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

func (s *GameServer) Cleanup() {
	close(s.done)
	s.world.Cleanup() // Cleanup từ world.go

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.listener.Close()
}

func generateID() string {
	return fmt.Sprintf("player_%d", time.Now().UnixNano())
}
func (s *GameServer) GetClientByID(clientID string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}
	return client, nil
}
