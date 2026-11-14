package models

import "github.com/google/uuid"

// Position on the game board
type Position struct {
	Col int `json:"col"`
	Row int `json:"row"`
}

// Resources
type Resources struct {
	Gold  float64 `json:"gold"`
	Wood  float64 `json:"wood"`
	Stone float64 `json:"stone"`
}

// BuildingType
type BuildingType string

const (
	BuildingCastle   BuildingType = "castle"
	BuildingFortress BuildingType = "fortress"
	BuildingVillage  BuildingType = "village"
)

// Building
type Building struct {
	ID       string       `json:"id"`
	Type     BuildingType `json:"type"`
	Owner    string       `json:"owner"`
	Position Position     `json:"position"`
	Level    int          `json:"level"`
	HP       int          `json:"hp"`
	MaxHP    int          `json:"maxHp"`
}

// UnitType
type UnitType string

const (
	UnitSwordsman    UnitType = "swordsman"
	UnitArcher       UnitType = "archer"
	UnitCavalry      UnitType = "cavalry"
	UnitCatapult     UnitType = "catapult"
	UnitWizard       UnitType = "wizard"
	UnitBeast        UnitType = "beast"
	UnitDragonKnight UnitType = "dragon_knight"
	UnitReaper       UnitType = "reaper"
)

// UnitState
type UnitState string

const (
	UnitIdle      UnitState = "idle"
	UnitMoving    UnitState = "moving"
	UnitAttacking UnitState = "attacking"
	UnitDead      UnitState = "dead"
)

// Unit
type Unit struct {
	ID             string     `json:"id"`
	Type           UnitType   `json:"type"`
	Owner          string     `json:"owner"`
	Position       Position   `json:"position"`
	TargetPosition *Position  `json:"targetPosition,omitempty"`
	TargetID       string     `json:"targetId,omitempty"`
	HP             int        `json:"hp"`
	MaxHP          int        `json:"maxHp"`
	ATK            int        `json:"atk"`
	State          UnitState  `json:"state"`
	LastAttackTime int64      `json:"lastAttackTime"`
}

// PlayerState
type PlayerState struct {
	Resources Resources  `json:"resources"`
	Buildings []Building `json:"buildings"`
	Units     []Unit     `json:"units"`
}

// GameState
type GameState struct {
	Tick          uint64                  `json:"tick"`
	TimeRemaining int                     `json:"timeRemaining"`
	Phase         string                  `json:"phase"` // waiting, playing, finished
	Winner        string                  `json:"winner,omitempty"`
	EndReason     string                  `json:"endReason,omitempty"`
	Players       map[string]*PlayerState `json:"players"`
}

// NewGameState creates initial game state
func NewGameState() *GameState {
	gs := &GameState{
		Tick:          0,
		TimeRemaining: 600,
		Phase:         "playing",
		Players:       make(map[string]*PlayerState),
	}

	for _, side := range []string{"player1", "player2"} {
		gs.Players[side] = &PlayerState{
			Resources: Resources{Gold: 50, Wood: 20, Stone: 20},
			Buildings: createInitialBuildings(side),
			Units:     []Unit{},
		}
	}

	return gs
}

func createInitialBuildings(side string) []Building {
	type pos struct{ col, row int }
	var positions map[string]pos

	if side == "player1" {
		positions = map[string]pos{
			"castle":    {0, 4},
			"fortress1": {0, 1},
			"fortress2": {0, 7},
			"village1":  {2, 1},
			"village2":  {2, 4},
			"village3":  {2, 7},
		}
	} else {
		positions = map[string]pos{
			"castle":    {11, 4},
			"fortress1": {11, 1},
			"fortress2": {11, 7},
			"village1":  {9, 1},
			"village2":  {9, 4},
			"village3":  {9, 7},
		}
	}

	return []Building{
		{ID: side + "_castle", Type: BuildingCastle, Owner: side, Position: Position{Col: positions["castle"].col, Row: positions["castle"].row}, Level: 1, HP: 500, MaxHP: 500},
		{ID: side + "_fortress1", Type: BuildingFortress, Owner: side, Position: Position{Col: positions["fortress1"].col, Row: positions["fortress1"].row}, Level: 1, HP: 300, MaxHP: 300},
		{ID: side + "_fortress2", Type: BuildingFortress, Owner: side, Position: Position{Col: positions["fortress2"].col, Row: positions["fortress2"].row}, Level: 1, HP: 300, MaxHP: 300},
		{ID: side + "_village1", Type: BuildingVillage, Owner: side, Position: Position{Col: positions["village1"].col, Row: positions["village1"].row}, Level: 1, HP: 150, MaxHP: 150},
		{ID: side + "_village2", Type: BuildingVillage, Owner: side, Position: Position{Col: positions["village2"].col, Row: positions["village2"].row}, Level: 1, HP: 150, MaxHP: 150},
		{ID: side + "_village3", Type: BuildingVillage, Owner: side, Position: Position{Col: positions["village3"].col, Row: positions["village3"].row}, Level: 1, HP: 150, MaxHP: 150},
	}
}

// GenerateID creates a new unique ID
func GenerateID() string {
	return uuid.New().String()
}
