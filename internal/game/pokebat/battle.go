package pokebat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

type BattleState string

const (
	BattleStateWaiting  BattleState = "waiting"
	BattleStateActive   BattleState = "active"
	BattleStateFinished BattleState = "finished"
)

type BattlePlayer struct {
	ID           string
	CurrentIndex int
	Team         []*models.Pokemon
	IsReady      bool
	HasSurrender bool
}

type Battle struct {
	ID           string
	Player1      *BattlePlayer
	Player2      *BattlePlayer
	State        BattleState
	CurrentTurn  string
	LastMoveTime time.Time
	StartTime    time.Time
	Logs         []string
	mu           sync.RWMutex
}

func NewBattle(id string, p1 *models.Player, p2 *models.Player) (*Battle, error) {
	// Validate players
	if p1.IsInBattle() || p2.IsInBattle() {
		return nil, fmt.Errorf("player already in battle")
	}

	// Setup battle players
	bp1, err := setupBattlePlayer(p1)
	if err != nil {
		return nil, err
	}
	bp2, err := setupBattlePlayer(p2)
	if err != nil {
		return nil, err
	}

	battle := &Battle{
		ID:           id,
		Player1:      bp1,
		Player2:      bp2,
		State:        BattleStateWaiting,
		LastMoveTime: time.Now(),
		StartTime:    time.Now(),
		Logs:         make([]string, 0),
	}

	return battle, nil
}

func setupBattlePlayer(p *models.Player) (*BattlePlayer, error) {
	team := p.GetBattleTeam()
	if len(team) != constants.MaxBattlePokemon {
		return nil, fmt.Errorf("invalid battle team size: expected %d, got %d",
			constants.MaxBattlePokemon, len(team))
	}

	battleTeam := make([]*models.Pokemon, len(team))
	for i, num := range team {
		pokemon, err := p.GetPokemon(num)
		if err != nil {
			return nil, fmt.Errorf("failed to get pokemon %s: %v", num, err)
		}
		if !pokemon.IsAlive() {
			return nil, fmt.Errorf("pokemon %s is not available for battle", num)
		}
		battleTeam[i] = pokemon
	}

	return &BattlePlayer{
		ID:           p.GetID(), // Truy cập trực tiếp ID từ PlayerData
		CurrentIndex: 0,
		Team:         battleTeam,
		IsReady:      false,
		HasSurrender: false,
	}, nil
}

// ExecuteMove - Thực hiện lượt đánh
func (b *Battle) ExecuteMove(playerID string, moveType string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.State != BattleStateActive {
		return fmt.Errorf("battle not active")
	}
	if b.CurrentTurn != playerID {
		return fmt.Errorf("not your turn")
	}
	if time.Since(b.StartTime).Seconds() > float64(constants.BattleTimeout) {
		b.State = BattleStateFinished
		return fmt.Errorf("battle timeout")
	}

	attacker, defender := b.getCurrentPokemon(playerID)
	if !attacker.IsAlive() || !defender.IsAlive() {
		return fmt.Errorf("invalid pokemon state")
	}

	// Validate move type
	if moveType != constants.NormalAttackType && moveType != constants.SpecialAttackType {
		return fmt.Errorf("invalid move type")
	}

	// Calculate and apply damage
	damage := b.calculateDamage(attacker, defender, moveType)
	defender.CurrentStats.HP -= damage
	b.logMove(playerID, attacker, defender, moveType, damage)

	// Check if defender fainted
	if !defender.IsAlive() {
		if err := b.handleFaintedPokemon(playerID); err != nil {
			return err
		}
	}

	b.switchTurn()
	b.LastMoveTime = time.Now()
	return nil
}

// StartBattle - Bắt đầu trận đấu
func (b *Battle) StartBattle() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.State != BattleStateWaiting {
		return fmt.Errorf("battle in invalid state")
	}
	if !b.Player1.IsReady || !b.Player2.IsReady {
		return fmt.Errorf("players not ready")
	}

	// Determine first turn based on Pokemon speed
	p1Speed := b.Player1.Team[0].CurrentStats.Speed
	p2Speed := b.Player2.Team[0].CurrentStats.Speed

	if p1Speed > p2Speed {
		b.CurrentTurn = b.Player1.ID
	} else if p2Speed > p1Speed {
		b.CurrentTurn = b.Player2.ID
	} else {
		// Random if speed equal
		if rand.Float64() < 0.5 {
			b.CurrentTurn = b.Player1.ID
		} else {
			b.CurrentTurn = b.Player2.ID
		}
	}

	b.State = BattleStateActive
	b.StartTime = time.Now()
	b.LastMoveTime = time.Now()
	return nil
}

// SetPlayerReady - Đánh dấu player sẵn sàng
func (b *Battle) SetPlayerReady(playerID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.State != BattleStateWaiting {
		return fmt.Errorf("battle already started")
	}

	if playerID == b.Player1.ID {
		b.Player1.IsReady = true
	} else if playerID == b.Player2.ID {
		b.Player2.IsReady = true
	} else {
		return fmt.Errorf("invalid player ID")
	}

	// Auto start if both ready
	if b.Player1.IsReady && b.Player2.IsReady {
		return b.StartBattle()
	}

	return nil
}

func (b *Battle) calculateDamage(attacker, defender *models.Pokemon, moveType string) int {
	if moveType == constants.NormalAttackType {
		return attacker.CurrentStats.Attack - defender.CurrentStats.Defense
	}

	// Special attack with type effectiveness
	baseDamage := attacker.CurrentStats.SpecialAtk
	maxMultiplier := 1.0

	for _, attackType := range attacker.GetTypes() {
		if effectiveness, exists := constants.TypeEffectiveness[attackType]; exists {
			for _, defenderType := range defender.GetTypes() {
				if multiplier, exists := effectiveness[defenderType]; exists {
					if multiplier > maxMultiplier {
						maxMultiplier = multiplier
					}
				}
			}
		}
	}

	return int(float64(baseDamage)*maxMultiplier) - defender.CurrentStats.SpecialDef
}

func (b *Battle) handleFaintedPokemon(playerID string) error {
	defender := b.getDefendingPlayer(playerID)

	// Find next available Pokemon
	for i := defender.CurrentIndex + 1; i < len(defender.Team); i++ {
		if defender.Team[i].IsAlive() {
			defender.CurrentIndex = i
			return nil
		}
	}

	// No more Pokemon available
	return b.endBattle(b.getAttackingPlayer(playerID).ID)
}

func (b *Battle) Surrender(playerID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.State != BattleStateActive {
		return fmt.Errorf("battle not active")
	}

	if playerID == b.Player1.ID {
		b.Player1.HasSurrender = true
		return b.endBattle(b.Player2.ID)
	} else {
		b.Player2.HasSurrender = true
		return b.endBattle(b.Player1.ID)
	}
}

func (b *Battle) endBattle(winnerID string) error {
	b.State = BattleStateFinished

	// Calculate total exp from losing team
	var totalExp int
	loserTeam := b.Player1.Team
	if b.Player1.ID == winnerID {
		loserTeam = b.Player2.Team
	}

	for _, pokemon := range loserTeam {
		totalExp += pokemon.GetExp()
	}

	// Distribute exp to winner's team
	expPerPokemon := totalExp / (3 * len(b.Player1.Team))
	winnerTeam := b.Player1.Team
	if b.Player2.ID == winnerID {
		winnerTeam = b.Player2.Team
	}

	for _, pokemon := range winnerTeam {
		if pokemon.IsAlive() {
			pokemon.AddExperience(expPerPokemon)
		}
	}

	return nil
}

// Helper methods...
