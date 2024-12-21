package terminal

import (
	"fmt"
	"strings"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/network/protocol"
)

type Display struct {
	width  int
	height int
}

func NewDisplay() *Display {
	return &Display{
		width:  constants.WorldWidth,
		height: constants.WorldHeight,
	}
}

// RenderWorldState - Hiển thị world grid 1000x1000
func (d *Display) RenderWorldState(state *protocol.WorldStatePayload) {
	// Clear screen trước khi render
	fmt.Print("\033[H\033[2J")

	// Hiển thị thông tin world
	fmt.Printf("=== POKEMON WORLD (%dx%d) ===\n", d.width, d.height)
	fmt.Printf("Players Online: %d | Active Pokemons: %d\n\n",
		len(state.Players), len(state.Pokemons))

	// Render viewport (20x20) quanh player's position
	viewportSize := 20
	playerPos := state.Players[state.CurrentPlayerID]

	startX := max(0, playerPos.X-viewportSize/2)
	startY := max(0, playerPos.Y-viewportSize/2)
	endX := min(d.width, startX+viewportSize)
	endY := min(d.height, startY+viewportSize)

	// Render grid với ASCII
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			// Default empty cell
			cell := "."

			// Check for player
			for id, pos := range state.Players {
				if pos.X == x && pos.Y == y {
					if id == state.CurrentPlayerID {
						cell = "@" // Current player
					} else {
						cell = "P" // Other players
					}
				}
			}

			// Check for pokemon
			for _, pokemon := range state.Pokemons {
				if pokemon.Position.X == x && pokemon.Position.Y == y {
					cell = "#"
				}
			}

			fmt.Printf("%s ", cell)
		}
		fmt.Println()
	}
}

// RenderBattleState - Hiển thị trạng thái battle
func (d *Display) RenderBattleState(battle *protocol.BattleStatePayload) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("=== POKEMON BATTLE ===")

	// Player 1 Pokemon
	p1 := battle.Player1Pokemon
	fmt.Printf("\nPlayer 1: %s (Level %d)\n", p1.Name, p1.Level)
	d.renderPokemonStats(p1)

	// Spacing
	fmt.Println("\n   VS")

	// Player 2 Pokemon
	p2 := battle.Player2Pokemon
	fmt.Printf("Player 2: %s (Level %d)\n", p2.Name, p2.Level)
	d.renderPokemonStats(p2)

	// Turn indicator
	fmt.Printf("\nCurrent Turn: Player %s\n", battle.CurrentTurn)
	fmt.Printf("Time Remaining: %s\n",
		time.Until(battle.TurnEndTime).Round(time.Second))

	// Battle log (last 5 entries)
	fmt.Println("\nBattle Log:")
	logs := battle.BattleLog
	if len(logs) > 5 {
		logs = logs[len(logs)-5:]
	}
	for _, log := range logs {
		fmt.Printf("> %s\n", log)
	}

	// Battle commands
	fmt.Println("\nCommands:")
	fmt.Println("1. Normal Attack")
	fmt.Println("2. Special Attack")
	fmt.Println("3. Switch Pokemon")
	fmt.Println("4. Surrender")
}

// renderPokemonStats - Helper để hiển thị chi tiết Pokemon
func (d *Display) renderPokemonStats(pokemon *models.Pokemon) {
	// HP bar
	hpPercentage := float64(pokemon.CurrentStats.HP) / float64(pokemon.BaseStats.HP)
	hpBar := strings.Repeat("=", int(20*hpPercentage))
	hpBar += strings.Repeat("-", 20-len(hpBar))

	fmt.Printf("HP: [%s] %d/%d\n", hpBar,
		pokemon.CurrentStats.HP, pokemon.BaseStats.HP)

	// Basic stats
	fmt.Printf("ATK: %d | DEF: %d\n",
		pokemon.CurrentStats.Attack, pokemon.CurrentStats.Defense)
	fmt.Printf("SP.ATK: %d | SP.DEF: %d\n",
		pokemon.CurrentStats.SpecialAtk, pokemon.CurrentStats.SpecialDef)
	fmt.Printf("Speed: %d | Types: %s\n",
		pokemon.CurrentStats.Speed, strings.Join(pokemon.Types, "/"))
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
