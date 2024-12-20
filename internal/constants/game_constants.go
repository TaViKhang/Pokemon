package constants

// Game World Constants
const (
	WorldWidth    = 1000 // Kích thước world theo yêu cầu
	WorldHeight   = 1000
	SpawnInterval = 60  // Spawn pokemon mỗi 60 giây
	DespawnTime   = 300 // Pokemon tự động despawn sau 5 phút
	SpawnCount    = 50  // Số lượng pokemon spawn mỗi đợt
)

// Player Constants
const (
	MaxPokemonInventory = 200 // Số lượng pokemon tối đa một player có thể sở hữu
	MovementSpeed       = 1   // Di chuyển 1 ô mỗi giây
	MaxBattlePokemon    = 3   // Số pokemon tối đa cho mỗi trận đấu
)

// Pokemon Stats Constants
const (
	DefaultEV = 0.5 // EV mặc định theo yêu cầu
	MinEV     = 0.5 // EV tối thiểu khi spawn
	MaxEV     = 1.0 // EV tối đa khi spawn
	MaxLevel  = 100 // Level tối đa của pokemon
)

// Experience Constants
const (
	ExpMultiplierPerLevel = 2 // Exp cần nhân đôi mỗi level
)

// Battle Constants
const (
	NormalAttackType  = "normal"  // Loại tấn công thường
	SpecialAttackType = "special" // Loại tấn công đặc biệt
	BattleTimeout     = 300       // Thời gian tối đa cho một trận đấu (giây)
)

// TypeEffectiveness - Bảng tương khắc chính thức giữa các type
var TypeEffectiveness = map[string]map[string]float64{
	"Normal": {
		"Rock":  0.5,
		"Ghost": 0.0,
		"Steel": 0.5,
	},
	"Fire": {
		"Fire":   0.5,
		"Water":  0.5,
		"Grass":  2.0,
		"Ice":    2.0,
		"Bug":    2.0,
		"Rock":   0.5,
		"Dragon": 0.5,
		"Steel":  2.0,
	},
	"Water": {
		"Fire":   2.0,
		"Water":  0.5,
		"Grass":  0.5,
		"Ground": 2.0,
		"Rock":   2.0,
		"Dragon": 0.5,
	},
	"Electric": {
		"Water":    2.0,
		"Electric": 0.5,
		"Grass":    0.5,
		"Ground":   0.0,
		"Flying":   2.0,
		"Dragon":   0.5,
	},
	"Grass": {
		"Fire":   0.5,
		"Water":  2.0,
		"Grass":  0.5,
		"Poison": 0.5,
		"Ground": 2.0,
		"Flying": 0.5,
		"Bug":    0.5,
		"Rock":   2.0,
		"Dragon": 0.5,
		"Steel":  0.5,
	},
	"Ice": {
		"Fire":   0.5,
		"Water":  0.5,
		"Grass":  2.0,
		"Ice":    0.5,
		"Ground": 2.0,
		"Flying": 2.0,
		"Dragon": 2.0,
		"Steel":  0.5,
	},
	"Fighting": {
		"Normal":  2.0,
		"Ice":     2.0,
		"Poison":  0.5,
		"Flying":  0.5,
		"Psychic": 0.5,
		"Bug":     0.5,
		"Rock":    2.0,
		"Ghost":   0.0,
		"Dark":    2.0,
		"Steel":   2.0,
		"Fairy":   0.5,
	},
	"Poison": {
		"Grass":  2.0,
		"Poison": 0.5,
		"Ground": 0.5,
		"Rock":   0.5,
		"Ghost":  0.5,
		"Steel":  0.0,
		"Fairy":  2.0,
	},
	"Ground": {
		"Fire":     2.0,
		"Electric": 2.0,
		"Grass":    0.5,
		"Poison":   2.0,
		"Flying":   0.0,
		"Bug":      0.5,
		"Rock":     2.0,
	},
	"Flying": {
		"Electric": 0.5,
		"Grass":    2.0,
		"Fighting": 2.0,
		"Bug":      2.0,
		"Rock":     0.5,
		"Steel":    0.5,
	},
	"Psychic": {
		"Fighting": 2.0,
		"Poison":   2.0,
		"Steel":    0.5,
		"Dark":     0.0,
	},
	"Bug": {
		"Fire":     0.5,
		"Grass":    2.0,
		"Fighting": 0.5,
		"Poison":   0.5,
		"Flying":   0.5,
		"Ghost":    0.5,
		"Steel":    0.5,
		"Fairy":    0.5,
	},
	"Rock": {
		"Fire":     2.0,
		"Ice":      2.0,
		"Fighting": 0.5,
		"Ground":   0.5,
		"Flying":   2.0,
		"Bug":      2.0,
		"Steel":    0.5,
	},
	"Ghost": {
		"Normal":  0.0,
		"Psychic": 2.0,
		"Ghost":   2.0,
		"Dark":    0.5,
	},
	"Dragon": {
		"Dragon": 2.0,
		"Steel":  0.5,
		"Fairy":  0.0,
	},
	"Dark": {
		"Fighting": 0.5,
		"Psychic":  2.0,
		"Ghost":    2.0,
		"Dark":     0.5,
		"Fairy":    0.5,
	},
	"Steel": {
		"Fire":     0.5,
		"Water":    0.5,
		"Electric": 0.5,
		"Ice":      2.0,
		"Rock":     2.0,
		"Steel":    0.5,
		"Fairy":    2.0,
	},
	"Fairy": {
		"Fire":     0.5,
		"Fighting": 2.0,
		"Poison":   0.5,
		"Dragon":   2.0,
		"Dark":     2.0,
		"Steel":    0.5,
	},
}

// Network Constants
const (
	TCPPort        = 8080
	MaxConnections = 100
	ReadTimeout    = 30 // Seconds
	WriteTimeout   = 30 // Seconds
	PingInterval   = 5  // Seconds
)

// Error Messages
const (
	ErrInventoryFull     = "pokemon inventory is full"
	ErrPokemonNotFound   = "pokemon not found"
	ErrInvalidMove       = "invalid movement"
	ErrBattleInProgress  = "battle already in progress"
	ErrInvalidBattleTeam = "invalid battle team selection"
	ErrInvalidLevel      = "invalid pokemon level"
	ErrInvalidExp        = "invalid experience points"
	ErrPokemonDestroyed  = "pokemon has been destroyed"
	ErrTypeMismatch      = "pokemon types do not match for exp transfer"
)

// Game States
type GameState int

const (
	StateIdle GameState = iota
	StateMoving
	StateBattling
	StateTrading
)

// Battle States
type BattleState int

const (
	BattleStateWaiting BattleState = iota
	BattleStateActive
	BattleStateFinished
)

// Movement Directions
type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
)

// File Paths
const (
	PokedexPath        = "data/pokedex.json"
	PlayerInventoryDir = "data/players/"
)
