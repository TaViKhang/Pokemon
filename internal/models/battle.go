package models

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
)

// Battle - Core model for a battle
type Battle struct {
	ID            string             `json:"id"`
	Player1       *Player            `json:"player1"`
	Player2       *Player            `json:"player2"`
	Player1Team   []string           `json:"player1_team"`
	Player2Team   []string           `json:"player2_team"`
	ActivePokemon map[string]string  `json:"active_pokemon"` // player ID -> pokemon number
	Turn          string             `json:"turn"`          // ID of the player whose turn it is
	StartTime     time.Time          `json:"start_time"`
	EndTime       *time.Time         `json:"end_time,omitempty"`
	IsFinished    bool               `json:"is_finished"`
	mu            sync.RWMutex
}

// NewBattle - Initializes a new battle
func NewBattle(id string, player1, player2 *Player) (*Battle, error) {
	if len(player1.GetBattleTeam()) != constants.MaxBattlePokemon || len(player2.GetBattleTeam()) != constants.MaxBattlePokemon {
		return nil, errors.New("both players must have a full battle team")
	}

	battle := &Battle{
		ID:            id,
		Player1:       player1,
		Player2:       player2,
		Player1Team:   player1.GetBattleTeam(),
		Player2Team:   player2.GetBattleTeam(),
		ActivePokemon: make(map[string]string),
		Turn:          player1.GetID(),
		StartTime:     time.Now(),
		IsFinished:    false,
	}

	battle.ActivePokemon[player1.GetID()] = battle.Player1Team[0]
	battle.ActivePokemon[player2.GetID()] = battle.Player2Team[0]

	return battle, nil
}

// GetActivePokemon - Retrieves the active Pokemon for a player
func (b *Battle) GetActivePokemon(playerID string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pokemon, exists := b.ActivePokemon[playerID]
	if !exists {
		return "", fmt.Errorf("no active pokemon for player %s", playerID)
	}
	return pokemon, nil
}

// SwitchPokemon - Switches the active Pokemon for a player
func (b *Battle) SwitchPokemon(playerID, pokemonNumber string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.IsFinished {
		return errors.New("cannot switch pokemon in a finished battle")
	}

	var team []string
	if playerID == b.Player1.GetID() {
		team = b.Player1Team
	} else if playerID == b.Player2.GetID() {
		team = b.Player2Team
	} else {
		return errors.New("invalid player ID")
	}

	found := false
	for _, num := range team {
		if num == pokemonNumber {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("pokemon %s is not in the team of player %s", pokemonNumber, playerID)
	}

	b.ActivePokemon[playerID] = pokemonNumber
	return nil
}

// PerformTurn - Processes a turn in the battle
func (b *Battle) PerformTurn(playerID string, action func(*Battle) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.IsFinished {
		return errors.New("cannot perform a turn in a finished battle")
	}

	if b.Turn != playerID {
		return errors.New("it's not your turn")
	}

	if err := action(b); err != nil {
		return fmt.Errorf("action failed: %v", err)
	}

	// Switch turn to the other player
	if b.Turn == b.Player1.GetID() {
		b.Turn = b.Player2.GetID()
	} else {
		b.Turn = b.Player1.GetID()
	}

	return nil
}

// EndBattle - Ends the battle
func (b *Battle) EndBattle(winnerID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.IsFinished {
		return errors.New("battle is already finished")
	}

	endTime := time.Now()
	b.EndTime = &endTime
	b.IsFinished = true

	// Clear the players' battle status
	b.Player1.mu.Lock()
	b.Player1.data.CurrentBattle = ""
	b.Player1.mu.Unlock()

	b.Player2.mu.Lock()
	b.Player2.data.CurrentBattle = ""
	b.Player2.mu.Unlock()

	return nil
}

// IsPlayerInBattle - Checks if a player is part of the battle
func (b *Battle) IsPlayerInBattle(playerID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return playerID == b.Player1.GetID() || playerID == b.Player2.GetID()
}
