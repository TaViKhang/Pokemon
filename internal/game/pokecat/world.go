package pokecat

import (
	"fmt"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

// Cell - Đại diện cho một ô trong world grid
type Cell struct {
	Players map[string]*models.Player
	Pokemon map[string]*models.Pokemon
	mu      sync.RWMutex
}

// Grid - Core model cho world grid system
type Grid struct {
	width     int
	height    int
	cells     [][]*Cell
	mu        sync.RWMutex
	spawnTick *time.Ticker
	done      chan struct{}
}

func NewGrid() *Grid {
	grid := &Grid{
		width:  constants.WorldWidth,
		height: constants.WorldHeight,
		cells:  make([][]*Cell, constants.WorldHeight),
		done:   make(chan struct{}),
	}

	// Khởi tạo cells
	for i := range grid.cells {
		grid.cells[i] = make([]*Cell, constants.WorldWidth)
		for j := range grid.cells[i] {
			grid.cells[i][j] = &Cell{
				Players: make(map[string]*models.Player),
				Pokemon: make(map[string]*models.Pokemon),
			}
		}
	}

	// Bắt đầu spawn routine
	grid.startSpawnRoutine()
	return grid
}

func (g *Grid) GetCell(x, y int) (*Cell, error) {
	if !g.isValidPosition(x, y) {
		return nil, fmt.Errorf("invalid position: (%d,%d)", x, y)
	}
	return g.cells[y][x], nil
}

// AddPlayer - Thêm player vào world
func (g *Grid) AddPlayer(player *models.Player) error {
	pos := player.GetPosition()
	cell, err := g.GetCell(pos.X, pos.Y)
	if err != nil {
		return err
	}

	cell.mu.Lock()
	cell.Players[player.GetID()] = player
	cell.mu.Unlock()
	return nil
}

// RemovePlayer - Xóa player khỏi world
func (g *Grid) RemovePlayer(player *models.Player) error {
	pos := player.GetPosition()
	cell, err := g.GetCell(pos.X, pos.Y)
	if err != nil {
		return err
	}

	cell.mu.Lock()
	delete(cell.Players, player.GetID())
	cell.mu.Unlock()
	return nil
}

// GetNearbyPokemons - Lấy danh sách Pokemon trong phạm vi có thể bắt
func (g *Grid) GetNearbyPokemons(x, y int) map[string]*models.Pokemon {
	result := make(map[string]*models.Pokemon)
	cell, err := g.GetCell(x, y)
	if err != nil {
		return result
	}

	cell.mu.RLock()
	for num, pokemon := range cell.Pokemon {
		result[num] = pokemon
	}
	cell.mu.RUnlock()
	return result
}

// CatchPokemon - Bắt Pokemon tại vị trí chỉ định
func (g *Grid) CatchPokemon(x, y int, pokemonNumber string) (*models.Pokemon, error) {
	cell, err := g.GetCell(x, y)
	if err != nil {
		return nil, err
	}

	cell.mu.Lock()
	defer cell.mu.Unlock()

	pokemon, exists := cell.Pokemon[pokemonNumber]
	if !exists {
		return nil, fmt.Errorf("pokemon not found at position (%d,%d)", x, y)
	}

	// Xóa Pokemon khỏi world sau khi bắt
	delete(cell.Pokemon, pokemonNumber)
	return pokemon, nil
}

func (g *Grid) MovePlayer(player *models.Player, newX, newY int) error {
	if !g.isValidPosition(newX, newY) {
		return fmt.Errorf("invalid position: (%d,%d)", newX, newY)
	}

	oldPos := player.GetPosition()
	oldCell, _ := g.GetCell(oldPos.X, oldPos.Y)
	newCell, _ := g.GetCell(newX, newY)

	// Lock cells theo thứ tự để tránh deadlock
	if oldPos.Y < newY || (oldPos.Y == newY && oldPos.X < newX) {
		oldCell.mu.Lock()
		newCell.mu.Lock()
	} else {
		newCell.mu.Lock()
		oldCell.mu.Lock()
	}
	defer oldCell.mu.Unlock()
	defer newCell.mu.Unlock()

	// Update cells
	delete(oldCell.Players, player.GetID())
	newCell.Players[player.GetID()] = player

	return nil
}
