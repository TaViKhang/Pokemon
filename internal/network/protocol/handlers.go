package protocol

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/game/pokebat"
)

// PlayerPosition represents a player's position in the game world
type PlayerPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type CommandType string
type Command struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

const (
	CmdMove      CommandType = "move"
	CmdCatch     CommandType = "catch"
	CmdBattle    CommandType = "battle"
	CmdHeartbeat CommandType = "heartbeat"
	   CmdWorldState CommandType = "world_state"
)
// PokemonPosition represents a Pokemon's position in the game world
type PokemonPosition struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Pokemon string `json:"pokemon"`
}

type GameServer interface {
	GetBattle(battleID string) (*pokebat.Battle, error)
	ExecuteBattleAttack(payload BattleAttackPayload) error
	ExecutePokemonSwitch(payload BattleSwitchPayload) error
	BroadcastBattleState(battleID string) error
}

type MessageHandler struct {
	server GameServer
}

func NewMessageHandler(server GameServer) *MessageHandler {
	return &MessageHandler{server: server}
}

func (h *MessageHandler) HandleBattleMessage(playerID string, msg Message) error {
	switch msg.Type {
	case MsgBattleAttack:
		return h.handleBattleAttack(playerID, msg.Payload)
	case MsgBattleSwitch:
		return h.handleBattleSwitch(playerID, msg.Payload)
	case MsgBattleSurrender:
		return h.handleBattleSurrender(playerID, msg.Payload)
	default:
		return fmt.Errorf("unknown battle message type: %s", msg.Type)
	}
}

func (h *MessageHandler) handleBattleAttack(playerID string, payload json.RawMessage) error {
	var data BattleAttackPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	battle, err := h.server.GetBattle(data.BattleID)
	if err != nil {
		return err
	}

	moveType := constants.NormalAttackType
	if data.IsSpecial {
		moveType = constants.SpecialAttackType
	}
	return battle.ExecuteMove(playerID, moveType)
}

func (h *MessageHandler) handleBattleSwitch(playerID string, payload json.RawMessage) error {
	var data BattleSwitchPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	battle, err := h.server.GetBattle(data.BattleID)
	if err != nil {
		return err
	}

	if data.IsForced {
		return battle.HandleFaintedPokemon(playerID)
	}

	return fmt.Errorf("voluntary switch not implemented")
}

func (h *MessageHandler) handleBattleSurrender(playerID string, payload json.RawMessage) error {
	var data struct {
		BattleID string `json:"battle_id"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	battle, err := h.server.GetBattle(data.BattleID)
	if err != nil {
		return err
	}

	return battle.Surrender(playerID)
}

// ReadPayload reads a message from the connection and unmarshals it into the provided target structure.
func ReadPayload(conn net.Conn, target interface{}) error {
	buf := make([]byte, 4096)

	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read from connection: %w", err)
	}

	if n == 0 {
		return fmt.Errorf("no data received")
	}

	fmt.Printf("Received %d bytes of data\n", n)

	if err := json.Unmarshal(buf[:n], target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// SendCommand encodes and sends a command with type and payload to the server.
func SendCommand(conn net.Conn, cmdType CommandType, payload interface{}) error {
    // Marshal the payload
    cmdPayload, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    // Construct the command object
    command := Command{
        Type:    cmdType,
        Payload: cmdPayload,
    }

    // Marshal the command
    data, err := json.Marshal(command)
    if err != nil {
        return fmt.Errorf("failed to marshal command: %w", err)
    }

    // Send the command to the server
    _, err = conn.Write(data)
    if err != nil {
        return fmt.Errorf("failed to send command: %w", err)
    }

    return nil
}

// GetPlayerPositions returns a map of player positions (ID -> PlayerPosition)
func GetPlayerPositions() map[string]PlayerPosition {
	// You can retrieve player positions from your game state here.
	// For now, we will return an empty map or sample data.
	playerPositions := make(map[string]PlayerPosition)
	// Example of adding a player position
	playerPositions["player1"] = PlayerPosition{X: 10, Y: 20}
	playerPositions["player2"] = PlayerPosition{X: 15, Y: 25}

	return playerPositions
}

// GetPokemonPositions returns a list of pokemon positions
func GetPokemonPositions() []PokemonPosition {
	// This should be based on your game logic to fetch Pokémon positions
	// Here we return sample data
	pokemonPositions := []PokemonPosition{
		{X: 30, Y: 40, Pokemon: "Pikachu"},
		{X: 50, Y: 60, Pokemon: "Charizard"},
	}

	return pokemonPositions
}
