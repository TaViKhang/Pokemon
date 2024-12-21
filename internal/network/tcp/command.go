package tcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/game/pokebat"
)

type CommandType string

const (
	CmdMove      CommandType = "move"
	CmdCatch     CommandType = "catch"
	CmdBattle    CommandType = "battle"
	CmdHeartbeat CommandType = "heartbeat"
)

type Command struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (s *GameServer) handleCommand(client *Client, data []byte) error {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return err
	}

	switch cmd.Type {
	case CmdMove:
		return s.handleMoveCommand(client, cmd.Payload)
	case CmdCatch:
		return s.handleCatchCommand(client, cmd.Payload)
	case CmdBattle:
		return s.handleBattleCommand(client, cmd.Payload)
	case CmdHeartbeat:
		return s.handleHeartbeat(client, cmd.Payload)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

func (s *GameServer) handleMoveCommand(client *Client, payload json.RawMessage) error {
	var moveData struct {
		Direction string `json:"direction"`
	}
	if err := json.Unmarshal(payload, &moveData); err != nil {
		return err
	}

	// Tích hợp với world.go
	pos := client.Player.GetPosition()
	newX, newY := pos.X, pos.Y

	switch moveData.Direction {
	case "up":
		newY--
	case "down":
		newY++
	case "left":
		newX--
	case "right":
		newX++
	}

	return s.world.MovePlayer(client.Player, newX, newY)
}

func (s *GameServer) handleCatchCommand(client *Client, payload json.RawMessage) error {
	var catchData struct {
		PokemonNumber string `json:"pokemon_number"`
	}
	if err := json.Unmarshal(payload, &catchData); err != nil {
		return err
	}

	pos := client.Player.GetPosition()
	pokemon, err := s.world.CatchPokemon(pos.X, pos.Y, catchData.PokemonNumber)
	if err != nil {
		return err
	}

	return client.Player.AddPokemon(pokemon)
}

func (s *GameServer) handleBattleCommand(client *Client, payload json.RawMessage) error {
	var battleData struct {
		OpponentID string `json:"opponent_id"`
	}
	if err := json.Unmarshal(payload, &battleData); err != nil {
		return err
	}

	s.mu.RLock()
	opponent, exists := s.clients[battleData.OpponentID]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("opponent not found")
	}

	battle, err := pokebat.NewBattle(
		fmt.Sprintf("battle_%d", time.Now().UnixNano()),
		client.Player,
		opponent.Player,
	)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.battles[battle.ID] = battle
	s.mu.Unlock()

	return nil
}

func (s *GameServer) handleHeartbeat(client *Client, payload json.RawMessage) error {
	var heartbeat HeartbeatMessage
	if err := json.Unmarshal(payload, &heartbeat); err != nil {
		return err
	}

	client.LastPing = time.Now()
	return nil
}
