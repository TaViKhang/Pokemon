package tcp

import "fmt"

type GameError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *GameError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

const (
	ErrInvalidCommand = iota + 1000
	ErrBattleNotFound
	ErrPlayerNotFound
	ErrInvalidMove
	ErrPokemonNotFound
	ErrBattleInProgress
	ErrInvalidBattleState
	ErrHeartbeatTimeout
)

var errorMessages = map[int]string{
	ErrInvalidCommand:     "invalid command format",
	ErrBattleNotFound:     "battle not found",
	ErrPlayerNotFound:     "player not found",
	ErrInvalidMove:        "invalid move",
	ErrPokemonNotFound:    "pokemon not found",
	ErrBattleInProgress:   "battle already in progress",
	ErrInvalidBattleState: "invalid battle state",
	ErrHeartbeatTimeout:   "client heartbeat timeout",
}

func NewGameError(code int) error {
	msg, ok := errorMessages[code]
	if !ok {
		msg = "unknown error"
	}
	return &GameError{Code: code, Message: msg}
}
