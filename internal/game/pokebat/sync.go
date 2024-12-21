package pokebat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
)

type BattleSync struct {
	battle      *Battle
	actionQueue chan BattleAction
	turnTimeout time.Duration
	turnEndTime time.Time
	currentTurn string
	mu          sync.RWMutex
}

type BattleAction struct {
	PlayerID  string
	Type      string // "attack", "switch", "surrender"
	IsSpecial bool   // Cho attack
	Target    string // Cho switch
}

func NewBattleSync(battle *Battle) *BattleSync {
	return &BattleSync{
		battle:      battle,
		actionQueue: make(chan BattleAction, 10),
		turnTimeout: time.Second * constants.BattleTimeout,
	}
}

func (bs *BattleSync) Start() error {
	// Xác định first turn dựa trên speed theo pokemon.txt
	p1Speed := bs.battle.Player1.Team[0].GetBattleStats().Speed
	p2Speed := bs.battle.Player2.Team[0].GetBattleStats().Speed

	if p1Speed > p2Speed {
		bs.currentTurn = bs.battle.Player1.ID
	} else if p2Speed > p1Speed {
		bs.currentTurn = bs.battle.Player2.ID
	} else {
		// Random nếu speed bằng nhau
		if rand.Float64() < 0.5 {
			bs.currentTurn = bs.battle.Player1.ID
		} else {
			bs.currentTurn = bs.battle.Player2.ID
		}
	}

	bs.turnEndTime = time.Now().Add(bs.turnTimeout)
	go bs.processTurns()
	return nil
}

func (bs *BattleSync) processTurns() {
	for {
		select {
		case action := <-bs.actionQueue:
			bs.mu.Lock()
			if err := bs.validateAndExecuteAction(action); err != nil {
				bs.battle.AddLog(fmt.Sprintf("Error: %v", err))
			}
			bs.mu.Unlock()

		case <-time.After(time.Until(bs.turnEndTime)):
			bs.mu.Lock()
			bs.handleTurnTimeout()
			bs.mu.Unlock()
		}

		if bs.battle.State == BattleStateFinished {
			break
		}
	}
}

func (bs *BattleSync) validateAndExecuteAction(action BattleAction) error {
	// Validate turn
	if action.PlayerID != bs.currentTurn {
		return fmt.Errorf("not your turn")
	}

	switch action.Type {
	case "attack":
		return bs.executeAttack(action)
	case "switch":
		return bs.executeSwitch(action)
	case "surrender":
		return bs.executeSurrender(action)
	default:
		return fmt.Errorf("invalid action type")
	}
}

func (bs *BattleSync) executeAttack(action BattleAction) error {
	attacker, defender := bs.battle.getCurrentPokemon(action.PlayerID)

	var damage int
	if action.IsSpecial {
		// Sử dụng lại logic từ Battle.calculateDamage()
		damage = bs.battle.calculateDamage(attacker, defender, constants.SpecialAttackType)
	} else {
		damage = bs.battle.calculateDamage(attacker, defender, constants.NormalAttackType)
	}

	if damage < 0 {
		damage = 0
	}

	defender.CurrentStats.HP -= damage

	// Check fainted
	if defender.CurrentStats.HP <= 0 {
		defender.CurrentStats.HP = 0
		if err := bs.battle.HandleFaintedPokemon(bs.battle.getOpponentID(action.PlayerID)); err != nil {
			return err
		}
	}

	bs.nextTurn()
	return nil
}

func (bs *BattleSync) executeSwitch(action BattleAction) error {
	player := bs.battle.getPlayer(action.PlayerID)

	// Validate target pokemon
	targetIdx := -1
	for i, p := range player.Team {
		if p.Number == action.Target {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 || !player.Team[targetIdx].IsAlive() {
		return fmt.Errorf("invalid switch target")
	}

	player.CurrentIndex = targetIdx
	bs.nextTurn() // End turn after switching theo pokemon.txt
	return nil
}

func (bs *BattleSync) executeSurrender(action BattleAction) error {
	winnerID := bs.battle.getOpponentID(action.PlayerID)

	// Calculate exp gain - 1/3 total exp từ losing team theo pokemon.txt
	loserExp := 0
	loser := bs.battle.getPlayer(action.PlayerID)
	for _, p := range loser.Team {
		loserExp += p.AccumulatedExp
	}
	expGain := loserExp / 3

	// Distribute exp cho winning team
	winner := bs.battle.getPlayer(winnerID)
	expPerPokemon := expGain / len(winner.Team)
	for _, p := range winner.Team {
		p.AddExperience(expPerPokemon)
	}

	bs.battle.State = BattleStateFinished
	return nil
}

func (bs *BattleSync) handleTurnTimeout() {
	// Tự động chuyển lượt khi hết thời gian
	bs.battle.AddLog(fmt.Sprintf("Player %s turn timeout", bs.currentTurn))
	bs.nextTurn()
}

func (bs *BattleSync) nextTurn() {
	if bs.currentTurn == bs.battle.Player1.ID {
		bs.currentTurn = bs.battle.Player2.ID
	} else {
		bs.currentTurn = bs.battle.Player1.ID
	}
	bs.turnEndTime = time.Now().Add(bs.turnTimeout)
}

// Helper methods...
