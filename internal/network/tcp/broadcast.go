package tcp

import (
	"encoding/json"
	"time"
)

type BroadcastType string

const (
	BroadcastPlayerJoin   BroadcastType = "player_join"
	BroadcastPlayerMove   BroadcastType = "player_move"
	BroadcastPokemonSpawn BroadcastType = "pokemon_spawn"
	BroadcastBattleStart  BroadcastType = "battle_start"
	BroadcastBattleEnd    BroadcastType = "battle_end"
)

type BroadcastMessage struct {
	Type      BroadcastType `json:"type"`
	Payload   interface{}   `json:"payload"`
	Timestamp time.Time     `json:"timestamp"`
}

func (s *GameServer) broadcast(excludeID string, msgType BroadcastType, payload interface{}) error {
	msg := BroadcastMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, client := range s.clients {
		if id != excludeID {
			select {
			case client.SendChan <- data:
			default:
				// Channel full, skip
			}
		}
	}
	return nil
}
