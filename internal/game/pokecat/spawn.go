package pokecat

import (
	"math/rand"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

func (g *Grid) startSpawnRoutine() {
	g.spawnTick = time.NewTicker(time.Duration(constants.SpawnInterval) * time.Second)

	go func() {
		for {
			select {
			case <-g.spawnTick.C:
				g.spawnPokemonWave()
			case <-g.done:
				g.spawnTick.Stop()
				return
			}
		}
	}()
}

func (g *Grid) spawnPokemonWave() {
	for i := 0; i < constants.SpawnCount; i++ {
		// Random position
		x := rand.Intn(g.width)
		y := rand.Intn(g.height)

		// Random level và EV
		level := rand.Intn(constants.MaxLevel) + 1
		ev := constants.MinEV + rand.Float64()*(constants.MaxEV-constants.MinEV)

		pokemon, err := models.NewRandomPokemon(level, ev)
		if err != nil {
			continue
		}

		cell, _ := g.GetCell(x, y)
		cell.mu.Lock()
		cell.Pokemon[pokemon.Number] = pokemon
		cell.mu.Unlock()

		// Schedule despawn
		go g.scheduleDespawn(x, y, pokemon.Number)
	}
}

func (g *Grid) scheduleDespawn(x, y int, pokemonNumber string) {
	time.Sleep(time.Duration(constants.DespawnTime) * time.Second)

	cell, err := g.GetCell(x, y)
	if err != nil {
		return
	}

	cell.mu.Lock()
	delete(cell.Pokemon, pokemonNumber)
	cell.mu.Unlock()
}

func (g *Grid) isValidPosition(x, y int) bool {
	return x >= 0 && x < g.width && y >= 0 && y < g.height
}

func (g *Grid) Cleanup() {
	close(g.done)
}
