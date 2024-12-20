package pokebat

import (
	"fmt"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

func (b *Battle) getCurrentPokemon(playerID string) (*models.Pokemon, *models.Pokemon) {
	var attacker, defender *models.Pokemon
	if playerID == b.Player1.ID {
		attacker = b.Player1.Team[b.Player1.CurrentIndex]
		defender = b.Player2.Team[b.Player2.CurrentIndex]
	} else {
		attacker = b.Player2.Team[b.Player2.CurrentIndex]
		defender = b.Player1.Team[b.Player1.CurrentIndex]
	}
	return attacker, defender
}

func (b *Battle) getAttackingPlayer(playerID string) *BattlePlayer {
	if playerID == b.Player1.ID {
		return b.Player1
	}
	return b.Player2
}

func (b *Battle) getDefendingPlayer(playerID string) *BattlePlayer {
	if playerID == b.Player1.ID {
		return b.Player2
	}
	return b.Player1
}

func (b *Battle) switchTurn() {
	if b.CurrentTurn == b.Player1.ID {
		b.CurrentTurn = b.Player2.ID
	} else {
		b.CurrentTurn = b.Player1.ID
	}
	b.LastMoveTime = time.Now()
}

func (b *Battle) logMove(playerID string, attacker, defender *models.Pokemon, moveType string, damage int) {
	log := fmt.Sprintf("%s's %s used %s attack on %s's %s for %d damage",
		playerID,
		attacker.Name,
		moveType,
		b.getDefendingPlayer(playerID).ID,
		defender.Name,
		damage)
	b.Logs = append(b.Logs, log)
}
