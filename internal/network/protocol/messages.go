package protocol

import (
	"encoding/json"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

// Base message structure
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MessageType string

const (
	// World updates
	MsgWorldState     MessageType = "world_state"
	MsgPokemonSpawn   MessageType = "pokemon_spawn"
	MsgPokemonDespawn MessageType = "pokemon_despawn"

	// Battle messages
	MsgBattleStart  MessageType = "battle_start"
	MsgBattleTurn   MessageType = "battle_turn"
	MsgBattleAction MessageType = "battle_action"
	MsgBattleEnd    MessageType = "battle_end"

	// Player actions
	MsgPlayerMove  MessageType = "player_move"
	MsgPlayerCatch MessageType = "player_catch"
	MsgAutoMode    MessageType = "auto_mode"

	// Battle specific
	MsgBattleAttack    MessageType = "battle_attack"
	MsgBattleSwitch    MessageType = "battle_switch"
	MsgBattleSurrender MessageType = "battle_surrender"
	MsgBattleResult    MessageType = "battle_result"

	// Pokemon related
	MsgPokemonLevelUp MessageType = "pokemon_level_up"
	MsgPokemonDestroy MessageType = "pokemon_destroy"
	MsgExpTransfer    MessageType = "exp_transfer"
)

// World update payloads
type WorldStatePayload struct {
	GridSize        [2]int                    `json:"grid_size"`
	Players         map[string]Position       `json:"players"`
	Pokemons        map[string]models.Pokemon `json:"pokemons"`
	Timestamp       time.Time                 `json:"timestamp"`
	CurrentPlayerID string                    `json:"current_player_id"`
}

type PokemonSpawnPayload struct {
	Number    string    `json:"number"`
	Position  Position  `json:"position"`
	Level     int       `json:"level"`
	EV        float64   `json:"ev"`
	DespawnAt time.Time `json:"despawn_at"`
}

// Battle payloads
type BattleStartPayload struct {
	BattleID  string   `json:"battle_id"`
	Player1   string   `json:"player1_id"`
	Player2   string   `json:"player2_id"`
	Team1     []string `json:"team1"` // 3 pokemons theo pokemon.txt
	Team2     []string `json:"team2"`
	FirstTurn string   `json:"first_turn"` // Dựa trên speed
}

type BattleActionPayload struct {
	BattleID string       `json:"battle_id"`
	PlayerID string       `json:"player_id"`
	Action   BattleAction `json:"action"`
	Target   string       `json:"target"`
}

type BattleAction string

const (
	ActionAttack    BattleAction = "attack" // Normal/Special theo pokemon.txt
	ActionSwitch    BattleAction = "switch" // Switch pokemon
	ActionSurrender BattleAction = "surrender"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Battle specific payloads
type BattleAttackPayload struct {
	BattleID     string  `json:"battle_id"`
	AttackerID   string  `json:"attacker_id"`
	DefenderID   string  `json:"defender_id"`
	IsSpecial    bool    `json:"is_special"` // Special hay Normal attack
	Damage       int     `json:"damage"`
	ElementalMul float64 `json:"elemental_mul"` // ELEMENTAL_MULTIPLY từ pokemon.txt
}

type BattleSwitchPayload struct {
	BattleID   string `json:"battle_id"`
	PlayerID   string `json:"player_id"`
	OldPokemon string `json:"old_pokemon"`
	NewPokemon string `json:"new_pokemon"`
	IsForced   bool   `json:"is_forced"` // True nếu pokemon died
}

type BattleResultPayload struct {
	BattleID  string           `json:"battle_id"`
	WinnerID  string           `json:"winner_id"`
	ExpGained int              `json:"exp_gained"` // 1/3 total exp từ losing team
	TeamExp   []PokemonExpGain `json:"team_exp"`
}

type PokemonExpGain struct {
	Number  string `json:"number"`
	OldExp  int    `json:"old_exp"`
	NewExp  int    `json:"new_exp"`
	Leveled bool   `json:"leveled"`
}

// BattleStatePayload - Thông tin trạng thái battle
type BattleStatePayload struct {
	BattleID       string          `json:"battle_id"`
	Player1Pokemon *models.Pokemon `json:"player1_pokemon"`
	Player2Pokemon *models.Pokemon `json:"player2_pokemon"`
	CurrentTurn    string          `json:"current_turn"`
	TurnEndTime    time.Time       `json:"turn_end_time"`
	BattleLog      []string        `json:"battle_log"`
}
