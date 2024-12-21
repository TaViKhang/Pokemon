package tcp

import (
	"sync"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/constants"
)

type SessionState struct {
	LastActivity time.Time
	AutoMode     struct {
		Active    bool
		Duration  time.Duration
		EndTime   time.Time
		Direction constants.Direction
	}
	CurrentBattle string // Battle ID
	Inventory     struct {
		PokemonCount int
		LastCatch    time.Time
	}
}

type SessionManager struct {
	sessions map[string]*SessionState
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionState),
	}
}

func (sm *SessionManager) CreateSession(playerID string) *SessionState {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &SessionState{
		LastActivity: time.Now(),
	}
	sm.sessions[playerID] = session
	return session
}

func (sm *SessionManager) GetSession(playerID string) (*SessionState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[playerID]
	return session, exists
}

func (sm *SessionManager) RemoveSession(playerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, playerID)
}
