package pokecat

import (
	"fmt"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/network/protocol"
)

// Cell - Represents a cell in the world grid
type Cell struct {
	Players map[string]*models.Player
	Pokemon map[string]*models.Pokemon
	mu      sync.RWMutex
}

// Grid - Core model for world grid system
type Grid struct {
	width     int
	height    int
	cells     [][]*Cell
	// mu        sync.RWMutex
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

	// Initialize cells
	for i := range grid.cells {
		grid.cells[i] = make([]*Cell, constants.WorldWidth)
		for j := range grid.cells[i] {
			grid.cells[i][j] = &Cell{
				Players: make(map[string]*models.Player),
				Pokemon: make(map[string]*models.Pokemon),
			}
		}
	}

	// Start spawn routine (if needed)
	grid.startSpawnRoutine()
	return grid
}

func (g *Grid) GetCell(x, y int) (*Cell, error) {
	if !g.isValidPosition(x, y) {
		return nil, fmt.Errorf("invalid position: (%d,%d)", x, y)
	}
	return g.cells[y][x], nil
}

// AddPlayer - Add a player to the world
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

// RemovePlayer - Remove a player from the world
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

// GetNearbyPokemons - Get the list of Pokemons in the catchable range
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

// CatchPokemon - Catch a Pokemon at the specified position
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

	// Remove Pokemon from the world after catching
	delete(cell.Pokemon, pokemonNumber)
	return pokemon, nil
}

// MovePlayer - Move a player to a new position
func (g *Grid) MovePlayer(player *models.Player, newX, newY int) error {
	if !g.isValidPosition(newX, newY) {
		return fmt.Errorf("invalid position: (%d,%d)", newX, newY)
	}

	oldPos := player.GetPosition()
	oldCell, _ := g.GetCell(oldPos.X, oldPos.Y)
	newCell, _ := g.GetCell(newX, newY)

	// Lock cells in order to avoid deadlock
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

// GetState - Get the world state, including all players and pokemons
func (g *Grid) GetState() protocol.WorldStatePayload {
	// Create a map for player positions
	playerPositions := make(map[string]protocol.Position)
	for _, cellRow := range g.cells {
		for _, cell := range cellRow {
			cell.mu.RLock()
			for _, player := range cell.Players {
				pos := player.GetPosition()
				playerPositions[player.GetID()] = protocol.Position{
					X: pos.X,
					Y: pos.Y,
				}
			}
			cell.mu.RUnlock()
		}
	}

	// Create a list for pokemon positions
	var pokemonPositions = make(map[string]models.Pokemon)  // Initialize map properly

	for _, cellRow := range g.cells {
		for _, cell := range cellRow {
			cell.mu.RLock()
			for _, pokemon := range cell.Pokemon {
				// Get the position of the Pokémon
				pos := pokemon.GetPosition()
	
				// Add the Pokémon to the map with its ID (e.g., using pokemon.Name or pokemon.Number as the key)
				pokemonPositions[pokemon.Name] = models.Pokemon{
					// You can include other fields from the Pokemon struct if needed
					Position: pos,
				}
			}
			cell.mu.RUnlock()
		}
	}

	// Assume "player_1" as the current player ID for now
	return protocol.WorldStatePayload{
		Players:        playerPositions,
		Pokemons:       pokemonPositions,
		CurrentPlayerID: "player_1", // This should be dynamic if needed
	}
}
