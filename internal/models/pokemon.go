package models

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
)

// Stats - Basic stats for a Pokemon
type Stats struct {
	HP         int `json:"hp"`
	Attack     int `json:"attack"`
	Defense    int `json:"defense"`
	SpecialAtk int `json:"sp_atk"`
	SpecialDef int `json:"sp_def"`
	Speed      int `json:"speed"`
	Total      int `json:"total"`
}

// PokemonPosition - Position of the Pokemon on the map
type PokemonPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Pokemon - Core model for each Pokemon
type Pokemon struct {
	FullName       string          `json:"full_name"`
	Name           string          `json:"name"`
	Number         string          `json:"number"`
	Position       PokemonPosition `json:"position"`
	Types          []string        `json:"type"`
	BaseStats      Stats           `json:"base_stats"`
	CurrentStats   Stats           `json:"current_stats"`
	Level          int             `json:"level"`
	AccumulatedExp int             `json:"accumulated_exp"`
	BaseExp        int             `json:"base_exp"`
	EV             float64         `json:"ev"`
	IsDestroyed    bool            `json:"is_destroyed"`
}

// NewPokemon - Create a new Pokemon from Pokedex data
func NewPokemon(pokedexData map[string]interface{}, level int, ev float64) (*Pokemon, error) {
	// Validate input
	if level < 1 || level > constants.MaxLevel {
		return nil, fmt.Errorf(constants.ErrInvalidLevel)
	}
	if ev < constants.MinEV || ev > constants.MaxEV {
		return nil, fmt.Errorf("invalid EV value")
	}

	// Parse types
	typeStr := pokedexData["type"].(string)
	types := strings.Split(typeStr, " ")

	// Create base stats
	baseStats := Stats{
		HP:         int(pokedexData["hp"].(float64)),
		Attack:     int(pokedexData["attack"].(float64)),
		Defense:    int(pokedexData["defense"].(float64)),
		SpecialAtk: int(pokedexData["sp_atk"].(float64)),
		SpecialDef: int(pokedexData["sp_def"].(float64)),
		Speed:      int(pokedexData["speed"].(float64)),
		Total:      int(pokedexData["total"].(float64)),
	}

	// Generate Pokemon object
	pokemon := &Pokemon{
		FullName:       pokedexData["full_name"].(string),
		Name:           pokedexData["name"].(string),
		Number:         pokedexData["number"].(string),
		Types:          types,
		BaseStats:      baseStats,
		Level:          level,
		AccumulatedExp: 0,
		BaseExp:        int(pokedexData["base_exp"].(float64)),
		EV:             ev,
		IsDestroyed:    false,
	}

	// Recalculate stats based on level and EV
	pokemon.recalculateStats()
	return pokemon, nil
}

// NewRandomPokemon - Create a random Pokemon with the given level and EV
func NewRandomPokemon(level int, ev float64) (*Pokemon, error) {
	// Validate input
	if level < 1 || level > constants.MaxLevel {
		return nil, fmt.Errorf("invalid level: %d", level)
	}
	if ev < constants.MinEV || ev > constants.MaxEV {
		return nil, fmt.Errorf("invalid EV: %f", ev)
	}

	// Load Pokedex data
	pokedexData, err := loadPokedexData()
	if err != nil {
		return nil, fmt.Errorf("failed to load pokedex: %v", err)
	}

	// Randomly select a Pokemon from Pokedex
	pokemons := make([]map[string]interface{}, 0)
	for _, data := range pokedexData {
		if pokemon, ok := data.(map[string]interface{}); ok {
			pokemons = append(pokemons, pokemon)
		}
	}

	if len(pokemons) == 0 {
		return nil, fmt.Errorf("no pokemon data available")
	}

	randomPokemon := pokemons[rand.Intn(len(pokemons))]
	return NewPokemon(randomPokemon, level, ev)
}

// loadPokedexData - Load data from pokedex.json
func loadPokedexData() (map[string]interface{}, error) {
	data, err := os.ReadFile("data/pokedex.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read pokedex.json: %v", err)
	}

	var pokedex map[string]interface{}
	if err := json.Unmarshal(data, &pokedex); err != nil {
		return nil, fmt.Errorf("failed to parse pokedex data: %v", err)
	}

	return pokedex, nil
}

// AddExperience - Add experience and level up if needed
func (p *Pokemon) AddExperience(exp int) (bool, error) {
	if p.IsDestroyed {
		return false, fmt.Errorf(constants.ErrPokemonDestroyed)
	}
	if exp < 0 {
		return false, fmt.Errorf(constants.ErrInvalidExp)
	}
	if p.Level >= constants.MaxLevel {
		return false, nil
	}

	p.AccumulatedExp += exp
	requiredExp := p.calculateRequiredExp()

	if p.AccumulatedExp >= requiredExp {
		p.Level++
		p.recalculateStats()
		return true, nil
	}
	return false, nil
}

// calculateRequiredExp - Calculate the required exp for the next level
func (p *Pokemon) calculateRequiredExp() int {
	return p.BaseExp * int(math.Pow(constants.ExpMultiplierPerLevel, float64(p.Level-1)))
}

// recalculateStats - Recalculate stats based on level and EV
func (p *Pokemon) recalculateStats() {
	multiplier := 1.0 + p.EV
	p.CurrentStats = Stats{
		HP:         int(float64(p.BaseStats.HP) * multiplier),
		Attack:     int(float64(p.BaseStats.Attack) * multiplier),
		Defense:    int(float64(p.BaseStats.Defense) * multiplier),
		SpecialAtk: int(float64(p.BaseStats.SpecialAtk) * multiplier),
		SpecialDef: int(float64(p.BaseStats.SpecialDef) * multiplier),
		Speed:      p.BaseStats.Speed, // Speed does not change with EV
		Total:      0,                 // Total will be recalculated
	}

	// Recalculate Total
	p.CurrentStats.Total = p.CurrentStats.HP + p.CurrentStats.Attack +
		p.CurrentStats.Defense + p.CurrentStats.SpecialAtk +
		p.CurrentStats.SpecialDef + p.CurrentStats.Speed
}

// TransferExpToSameType - Transfer experience to a Pokemon of the same type
func (p *Pokemon) TransferExpToSameType(target *Pokemon) error {
	if p.IsDestroyed {
		return fmt.Errorf(constants.ErrPokemonDestroyed)
	}
	if target == nil {
		return fmt.Errorf("invalid target pokemon")
	}
	if target.IsDestroyed {
		return fmt.Errorf("cannot transfer to destroyed pokemon")
	}
	if target.Level >= constants.MaxLevel {
		return fmt.Errorf("target pokemon already at max level")
	}
	if !p.hasSameType(target) {
		return fmt.Errorf(constants.ErrTypeMismatch)
	}

	target.AddExperience(p.AccumulatedExp)
	p.AccumulatedExp = 0
	return nil
}

// DestroyPokemon - Destroy the Pokemon after transferring exp
func (p *Pokemon) DestroyPokemon() error {
	if p.AccumulatedExp > 0 {
		return fmt.Errorf("cannot destroy pokemon with remaining exp")
	}
	p.IsDestroyed = true
	return nil
}

// hasSameType - Check if the Pokemon has the same type as another
func (p *Pokemon) hasSameType(other *Pokemon) bool {
	if len(p.Types) != len(other.Types) {
		return false
	}

	for _, t1 := range p.Types {
		found := false
		for _, t2 := range other.Types {
			if t1 == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetBattleStats - Get stats for battle system
func (p *Pokemon) GetBattleStats() Stats {
	return p.CurrentStats
}

// GetTypes - Get types for battle system
func (p *Pokemon) GetTypes() []string {
	return p.Types
}

// IsAlive - Check if the Pokemon is alive
func (p *Pokemon) IsAlive() bool {
	return !p.IsDestroyed && p.CurrentStats.HP > 0
}

// GetLevel - Get current level
func (p *Pokemon) GetLevel() int {
	return p.Level
}

// GetExp - Get current experience
func (p *Pokemon) GetExp() int {
	return p.AccumulatedExp
}

// GetRequiredExp - Get required experience for next level
func (p *Pokemon) GetRequiredExp() int {
	return p.calculateRequiredExp()
}

// GetPosition - Get the current position of the Pokemon
func (p *Pokemon) GetPosition() PokemonPosition {
	return p.Position
}

// GetName - Get the name of the Pokemon
func (p *Pokemon) GetName() string {
	return p.Name
}
