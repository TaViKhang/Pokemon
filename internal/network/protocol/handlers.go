package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/game/pokebat"
)

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
