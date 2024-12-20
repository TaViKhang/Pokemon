package models

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
)

// PlayerData - Cấu trúc dữ liệu để lưu vào file JSON
type PlayerData struct {
	ID              string              `json:"id"`
	PokemonList     map[string]*Pokemon `json:"pokemon_list"`
	Position        Position            `json:"position"`
	CurrentBattle   string              `json:"current_battle_id,omitempty"`
	BattleTeam      []string            `json:"battle_team,omitempty"`
	LastMoveTime    time.Time           `json:"last_move_time"`
	AutoModeEndTime time.Time           `json:"auto_mode_end_time,omitempty"`
	LastSaveTime    time.Time           `json:"last_save_time"`
}

// Position - Vị trí của player trong world
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Player - Core model cho mỗi player
type Player struct {
	data         PlayerData
	mu           sync.RWMutex
	stopAutoSave chan struct{}
	isOnline     bool
}

// NewPlayer - Tạo player mới
func NewPlayer(id string) *Player {
	player := &Player{
		data: PlayerData{
			ID:          id,
			PokemonList: make(map[string]*Pokemon),
			Position: Position{
				X: rand.Intn(constants.WorldWidth),
				Y: rand.Intn(constants.WorldHeight),
			},
			LastMoveTime: time.Now(),
			LastSaveTime: time.Now(),
		},
		stopAutoSave: make(chan struct{}),
		isOnline:     true,
	}

	go player.startAutoSave()
	return player
}

// GetID - Lấy ID của player
func (p *Player) GetID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ID
}

// AddPokemon - Thêm pokemon vào inventory
func (p *Player) AddPokemon(pokemon *Pokemon) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.data.PokemonList) >= constants.MaxPokemonInventory {
		return fmt.Errorf(constants.ErrInventoryFull)
	}

	// Deep copy của pokemon để tránh reference issues
	pokemonCopy := *pokemon
	p.data.PokemonList[pokemon.Number] = &pokemonCopy

	return p.saveToFile()
}

// GetPokemon - Lấy thông tin Pokemon theo number
func (p *Player) GetPokemon(number string) (*Pokemon, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pokemon, exists := p.data.PokemonList[number]
	if !exists {
		return nil, fmt.Errorf(constants.ErrPokemonNotFound)
	}

	// Return deep copy để tránh race conditions
	pokemonCopy := *pokemon
	return &pokemonCopy, nil
}

// SelectBattleTeam - Chọn team cho battle
func (p *Player) SelectBattleTeam(pokemonNumbers []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(pokemonNumbers) != constants.MaxBattlePokemon {
		return fmt.Errorf("invalid team size: expected %d, got %d",
			constants.MaxBattlePokemon, len(pokemonNumbers))
	}

	// Validate từng Pokemon
	for _, num := range pokemonNumbers {
		pokemon, exists := p.data.PokemonList[num]
		if !exists {
			return fmt.Errorf("pokemon %s not found in inventory", num)
		}
		if !pokemon.IsAlive() {
			return fmt.Errorf("pokemon %s is not available for battle", num)
		}
		if pokemon.IsDestroyed {
			return fmt.Errorf("pokemon %s has been destroyed", num)
		}
	}

	p.data.BattleTeam = make([]string, len(pokemonNumbers))
	copy(p.data.BattleTeam, pokemonNumbers)
	return p.saveToFile()
}

// Move - Di chuyển player
func (p *Player) Move(direction constants.Direction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isOnline {
		return fmt.Errorf("player is offline")
	}

	if time.Since(p.data.LastMoveTime).Seconds() < float64(1/constants.MovementSpeed) {
		return fmt.Errorf("movement too frequent")
	}

	newPos := p.data.Position
	switch direction {
	case constants.DirectionUp:
		newPos.Y = (newPos.Y - 1 + constants.WorldHeight) % constants.WorldHeight
	case constants.DirectionDown:
		newPos.Y = (newPos.Y + 1) % constants.WorldHeight
	case constants.DirectionLeft:
		newPos.X = (newPos.X - 1 + constants.WorldWidth) % constants.WorldWidth
	case constants.DirectionRight:
		newPos.X = (newPos.X + 1) % constants.WorldWidth
	default:
		return fmt.Errorf("invalid direction")
	}

	p.data.Position = newPos
	p.data.LastMoveTime = time.Now()
	return p.saveToFile()
}

// SaveToFile - Lưu player data vào file JSON
func (p *Player) saveToFile() error {
	if err := os.MkdirAll(constants.PlayerInventoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	filename := filepath.Join(constants.PlayerInventoryDir, fmt.Sprintf("%s.json", p.data.ID))

	// Cập nhật thời gian save
	p.data.LastSaveTime = time.Now()

	data, err := json.MarshalIndent(p.data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal player data: %v", err)
	}

	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %v", err)
	}

	// Atomic rename để đảm bảo tính toàn vẹn của file
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up
		return fmt.Errorf("failed to save player data: %v", err)
	}

	return nil
}

// LoadFromFile - Load player data từ file JSON
func LoadPlayer(id string) (*Player, error) {
	filename := filepath.Join(constants.PlayerInventoryDir, fmt.Sprintf("%s.json", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return NewPlayer(id), nil
		}
		return nil, fmt.Errorf("failed to read player data: %v", err)
	}

	var playerData PlayerData
	if err := json.Unmarshal(data, &playerData); err != nil {
		return nil, fmt.Errorf("failed to parse player data: %v", err)
	}

	player := &Player{
		data:         playerData,
		stopAutoSave: make(chan struct{}),
		isOnline:     true,
	}

	go player.startAutoSave()
	return player, nil
}

// startAutoSave - Auto-save routine
func (p *Player) startAutoSave() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			if p.isOnline {
				if err := p.saveToFile(); err != nil {
					log.Printf("Auto-save failed for player %s: %v", p.data.ID, err)
				}
			}
			p.mu.Unlock()
		case <-p.stopAutoSave:
			return
		}
	}
}

// Cleanup - Dọn dẹp resources
func (p *Player) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isOnline = false
	close(p.stopAutoSave)
	return p.saveToFile()
}

// GetPosition - Lấy vị trí hiện tại
func (p *Player) GetPosition() Position {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.Position
}

// GetBattleTeam - Lấy thông tin battle team
func (p *Player) GetBattleTeam() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	team := make([]string, len(p.data.BattleTeam))
	copy(team, p.data.BattleTeam)
	return team
}

// IsInBattle - Kiểm tra player có đang trong battle
func (p *Player) IsInBattle() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.CurrentBattle != ""
}
