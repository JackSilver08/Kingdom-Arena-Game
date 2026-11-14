package game

import (
	"encoding/json"
	"errors"
	"fmt"

	"kingdom-arena/pkg/models"
)

var (
	ErrMatchFinished     = errors.New("match is already finished")
	ErrInvalidPlayer     = errors.New("invalid player")
	ErrInvalidCommand    = errors.New("invalid command")
	ErrInvalidUnit       = errors.New("invalid unit type")
	ErrInvalidPosition   = errors.New("invalid deployment position")
	ErrOccupiedPosition  = errors.New("position is occupied")
	ErrInsufficientFunds = errors.New("insufficient resources")
	ErrUnitLocked        = errors.New("unit is locked by fortress level")
	ErrUnitCooldown      = errors.New("unit deployment is on cooldown")
	ErrBuildingNotOwned  = errors.New("building is not owned by player")
	ErrBuildingDestroyed = errors.New("building is destroyed")
	ErrBuildingMaxLevel  = errors.New("building is already at maximum level")
	ErrTargetNotFound    = errors.New("target unit not found")
	ErrInvalidTarget     = errors.New("target must be an enemy unit")
	ErrDuplicateCommand  = errors.New("duplicate command")
)

const (
	PlayerOne = "player1"
	PlayerTwo = "player2"
	TickRate  = 10
)

type Command struct {
	Type       string `json:"type"`
	CommandID  string `json:"commandId"`
	UnitType   string `json:"unitType,omitempty"`
	Col        int    `json:"col,omitempty"`
	Row        int    `json:"row,omitempty"`
	BuildingID string `json:"buildingId,omitempty"`
	UnitID     string `json:"unitId,omitempty"`
	TargetID   string `json:"targetId,omitempty"`
}

type Event struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerId,omitempty"`
	EntityID string `json:"entityId,omitempty"`
	TargetID string `json:"targetId,omitempty"`
	Tick     uint64 `json:"tick"`
	Value    int    `json:"value,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type damageHit struct {
	attackerID    string
	ownerID       string
	targetID      string
	damage        int
	knockback     bool
	eventType     string
	eventEntityID string
	eventTargetID string
}

type Engine struct {
	State *models.GameState

	tick                 uint64
	nextDeployTick       map[string]uint64
	nextMoveTick         map[string]uint64
	nextAttackTick       map[string]uint64
	nextFortressAttack   map[string]uint64
	movedSinceSpawn      map[string]bool
	firstAttackConsumed  map[string]bool
	lastDragonBreathTick map[string]uint64
	buildingDestroyed    map[string]bool
	spawnTick            map[string]uint64
	reaperRevealed       map[string]bool
	beastChargeReady     map[string]bool
	manualTargets        map[string]Position
	seenCommands         map[string]struct{}
	commandOrder         []string
	events               []Event
}

type EngineFactory struct{}

func NewEngineFactory() *EngineFactory { return &EngineFactory{} }

func (f *EngineFactory) CreateEngine() *Engine {
	return NewEngine(models.NewGameState())
}

// NewEngine creates an authoritative simulation around an initialized game state.
// An Engine is intended to have a single owner goroutine (the room loop); callers
// must not mutate State directly while the engine is running.
func NewEngine(state *models.GameState) *Engine {
	if state == nil {
		state = models.NewGameState()
	}
	return &Engine{
		State:                state,
		tick:                 state.Tick,
		nextDeployTick:       make(map[string]uint64),
		nextMoveTick:         make(map[string]uint64),
		nextAttackTick:       make(map[string]uint64),
		nextFortressAttack:   make(map[string]uint64),
		movedSinceSpawn:      make(map[string]bool),
		firstAttackConsumed:  make(map[string]bool),
		lastDragonBreathTick: make(map[string]uint64),
		buildingDestroyed:    make(map[string]bool),
		spawnTick:            make(map[string]uint64),
		reaperRevealed:       make(map[string]bool),
		beastChargeReady:     make(map[string]bool),
		manualTargets:        make(map[string]Position),
		seenCommands:         make(map[string]struct{}),
	}
}

// ApplyCommand validates and applies one player intention. All consequential
// state changes happen here, never in the browser client.
func (e *Engine) ApplyCommand(playerID string, command Command) error {
	if e.State == nil {
		return ErrInvalidCommand
	}
	if e.State.Phase != "playing" {
		return ErrMatchFinished
	}
	if !isPlayer(playerID) || e.State.Players[playerID] == nil {
		return ErrInvalidPlayer
	}
	if command.CommandID != "" {
		key := playerID + ":" + command.CommandID
		if _, exists := e.seenCommands[key]; exists {
			return ErrDuplicateCommand
		}
		// Remember only the most recent 2,048 commands per engine.
		e.seenCommands[key] = struct{}{}
		e.commandOrder = append(e.commandOrder, key)
		if len(e.commandOrder) > 2048 {
			delete(e.seenCommands, e.commandOrder[0])
			e.commandOrder = e.commandOrder[1:]
		}
	}

	var err error
	switch command.Type {
	case "DEPLOY_UNIT":
		err = e.deployUnit(playerID, command)
	case "UPGRADE", "UPGRADE_BUILDING":
		err = e.upgradeBuilding(playerID, command.BuildingID)
	case "SET_TARGET":
		err = e.setTarget(playerID, command.UnitID, command.TargetID)
	case "MOVE_TO":
		err = e.moveUnit(playerID, command)
	case "SURRENDER":
		e.finish(opponent(playerID), "surrender")
	default:
		err = ErrInvalidCommand
	}
	if err == nil {
		e.events = append(e.events, Event{Type: "COMMAND_ACCEPTED", PlayerID: playerID, Tick: e.tick})
	}
	return err
}

// Tick advances the simulation by one fixed 100ms step and returns events
// produced during that step. Network serialization is deliberately outside
// the domain engine.
func (e *Engine) Tick() []Event {
	if e.State == nil {
		return nil
	}
	if e.State.Phase != "playing" {
		return e.drainEvents()
	}
	e.tick++
	e.State.Tick = e.tick

	if e.tick%TickRate == 0 {
		e.updateEconomy()
		if e.State.TimeRemaining > 0 {
			e.State.TimeRemaining--
		}
	}

	e.updateMovement()
	e.updateCombat()
	e.updateFortresses()
	e.updateDragonBreaths()
	e.removeDeadUnits()
	e.evaluateVictory()
	return e.drainEvents()
}

// RunBotTurn performs one deliberately simple, deterministic bot decision.
// It is a testing opponent, not a competitive AI.
func (e *Engine) RunBotTurn(playerID string) error {
	if e.State == nil || e.State.Phase != "playing" {
		return ErrMatchFinished
	}
	if !isPlayer(playerID) || e.State.Players[playerID] == nil {
		return ErrInvalidPlayer
	}
	player := e.State.Players[playerID]
	laneCounts := [3]int{}
	for _, ownUnit := range player.Units {
		laneCounts[laneForRow(ownUnit.Position.Row)]++
	}
	lane := 0
	for i := 1; i < len(laneCounts); i++ {
		if laneCounts[i] < laneCounts[lane] {
			lane = i
		}
	}
	rows := [][]int{{1, 0, 2}, {4, 3, 5, 6}, {7, 8, 9}}[lane]
	cols := []int{4, 3, 1, 0, 2}
	if playerID == PlayerTwo {
		cols = []int{7, 8, 10, 11, 9}
	}
	spawn := Position{Col: -1, Row: -1}
	for _, candidateCol := range cols {
		for _, candidateRow := range rows {
			candidate := Position{Col: candidateCol, Row: candidateRow}
			if !e.isOccupied(candidate) {
				spawn = candidate
				break
			}
		}
		if spawn.Col >= 0 {
			break
		}
	}
	if spawn.Col < 0 {
		return nil
	}

	unitChoice := []string{"swordsman", "archer", "cavalry", "beast", "wizard", "catapult", "reaper", "dragon_knight"}
	for _, unitType := range unitChoice {
		def, ok := unitDefinitions[unitType]
		if !ok || !e.isUnlocked(playerID, def.UnlockLevel) || e.tick < e.nextDeployTick[playerID+":"+unitType] {
			continue
		}
		if canAfford(player.Resources, def.Cost) {
			return e.ApplyCommand(playerID, Command{
				Type: "DEPLOY_UNIT", CommandID: fmt.Sprintf("bot-%s-%d", playerID, e.tick),
				UnitType: unitType, Col: spawn.Col, Row: spawn.Row,
			})
		}
	}
	return nil // The bot waits when it cannot afford or deploy a unit.
}

// Snapshot returns an isolated JSON-compatible copy for transport/storage.
func (e *Engine) Snapshot() (*models.GameState, error) {
	if e.State == nil {
		return nil, ErrInvalidCommand
	}
	data, err := json.Marshal(e.State)
	if err != nil {
		return nil, err
	}
	var snapshot models.GameState
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// SnapshotFor filters hidden enemy Reapers from a player's client view. The
// authoritative snapshot remains available only to server-side code.
func (e *Engine) SnapshotFor(playerID string) (*models.GameState, error) {
	if !isPlayer(playerID) || e.State.Players[playerID] == nil {
		return nil, ErrInvalidPlayer
	}
	snapshot, err := e.Snapshot()
	if err != nil {
		return nil, err
	}
	enemyID := opponent(playerID)
	visible := snapshot.Players[enemyID].Units[:0]
	for _, unit := range snapshot.Players[enemyID].Units {
		if unit.Type == models.UnitReaper && !e.reaperIsRevealed(unit) {
			continue
		}
		visible = append(visible, unit)
	}
	snapshot.Players[enemyID].Units = visible
	return snapshot, nil
}

func (e *Engine) deployUnit(playerID string, command Command) error {
	def, ok := unitDefinitions[command.UnitType]
	if !ok {
		return ErrInvalidUnit
	}
	if !e.isUnlocked(playerID, def.UnlockLevel) {
		return ErrUnitLocked
	}
	if !validSpawnPosition(playerID, Position{Col: command.Col, Row: command.Row}) {
		return ErrInvalidPosition
	}
	position := Position{Col: command.Col, Row: command.Row}
	if e.isOccupied(position) {
		return ErrOccupiedPosition
	}
	cooldownKey := playerID + ":" + command.UnitType
	if e.tick < e.nextDeployTick[cooldownKey] {
		return ErrUnitCooldown
	}
	player := e.State.Players[playerID]
	if !canAfford(player.Resources, def.Cost) {
		return ErrInsufficientFunds
	}
	spend(&player.Resources, def.Cost)
	atkBonus, hpBonus := e.castleBonuses(playerID)
	unit := models.Unit{
		ID: models.GenerateID(), Type: models.UnitType(command.UnitType), Owner: playerID,
		Position: models.Position{Col: position.Col, Row: position.Row},
		HP:       int(float64(def.HP) * (1 + hpBonus)),
		MaxHP:    int(float64(def.HP) * (1 + hpBonus)),
		ATK:      int(float64(def.ATK) * (1 + atkBonus)), State: models.UnitIdle,
		LastAttackTime: 0,
	}
	player.Units = append(player.Units, unit)
	e.spawnTick[unit.ID] = e.tick
	e.lastDragonBreathTick[unit.ID] = e.tick
	e.nextDeployTick[cooldownKey] = e.tick + uint64(def.TrainSeconds*TickRate)
	e.events = append(e.events, Event{Type: "UNIT_SPAWNED", PlayerID: playerID, EntityID: unit.ID, Tick: e.tick})
	return nil
}

func (e *Engine) upgradeBuilding(playerID, buildingID string) error {
	player := e.State.Players[playerID]
	var building *models.Building
	for i := range player.Buildings {
		if player.Buildings[i].ID == buildingID {
			building = &player.Buildings[i]
			break
		}
	}
	if building == nil {
		return ErrBuildingNotOwned
	}
	if building.HP <= 0 {
		return ErrBuildingDestroyed
	}
	if building.Level >= 3 {
		return ErrBuildingMaxLevel
	}
	nextLevel := building.Level + 1
	cost := buildingUpgradeCost(string(building.Type), nextLevel)
	if !canAfford(player.Resources, cost) {
		return ErrInsufficientFunds
	}
	spend(&player.Resources, cost)
	oldMaxHP := building.MaxHP
	building.Level = nextLevel
	building.MaxHP = buildingMaxHP(string(building.Type), nextLevel)
	building.HP += building.MaxHP - oldMaxHP
	if building.HP > building.MaxHP {
		building.HP = building.MaxHP
	}
	if building.Type == models.BuildingCastle {
		e.applyCastleBonuses(playerID)
	}
	e.events = append(e.events, Event{Type: "BUILDING_UPGRADED", PlayerID: playerID, EntityID: buildingID, Tick: e.tick, Value: nextLevel})
	return nil
}

func (e *Engine) setTarget(playerID, unitID, targetID string) error {
	unit := e.findUnit(unitID)
	if unit == nil || unit.Owner != playerID {
		return ErrTargetNotFound
	}
	target := e.findUnit(targetID)
	if target == nil {
		return ErrTargetNotFound
	}
	if target.Owner == playerID || target.Type == models.UnitReaper && !e.reaperIsRevealed(*target) {
		return ErrInvalidTarget
	}
	unit.TargetID = targetID
	unit.TargetBuildingID = ""
	unit.TargetPosition = &models.Position{Col: target.Position.Col, Row: target.Position.Row}
	delete(e.manualTargets, unitID)
	unit.State = models.UnitMoving
	e.events = append(e.events, Event{Type: "TARGET_SET", PlayerID: playerID, EntityID: unitID, TargetID: targetID, Tick: e.tick})
	return nil
}

// moveUnit issues a right-click order. Empty cells become a move destination;
// enemy units and buildings become attack targets.
func (e *Engine) moveUnit(playerID string, command Command) error {
	unit := e.findUnit(command.UnitID)
	if unit == nil || unit.Owner != playerID {
		return ErrTargetNotFound
	}
	target := Position{Col: command.Col, Row: command.Row}
	if target.Col < 0 || target.Col >= 12 || target.Row < 0 || target.Row >= 10 {
		return ErrInvalidPosition
	}
	// Lane switching is intentionally disabled in the current rules baseline.
	if laneForRow(target.Row) != laneForRow(unit.Position.Row) {
		return ErrInvalidPosition
	}

	unit.TargetID = ""
	unit.TargetBuildingID = ""
	unit.TargetPosition = &models.Position{Col: target.Col, Row: target.Row}
	delete(e.manualTargets, unit.ID)

	if enemy := e.findUnitAt(target); enemy != nil {
		if enemy.Owner == playerID || enemy.Type == models.UnitReaper && !e.reaperIsRevealed(*enemy) {
			return ErrInvalidTarget
		}
		unit.TargetID = enemy.ID
	} else if building := e.findBuildingAt(target); building != nil {
		if building.Owner == playerID {
			return ErrInvalidTarget
		}
		unit.TargetBuildingID = building.ID
	} else {
		e.manualTargets[unit.ID] = target
	}
	unit.State = models.UnitMoving
	e.events = append(e.events, Event{Type: "MOVE_ORDERED", PlayerID: playerID, EntityID: unit.ID, Tick: e.tick})
	return nil
}

func (e *Engine) findUnitAt(position Position) *models.Unit {
	for _, player := range e.State.Players {
		for i := range player.Units {
			unit := &player.Units[i]
			if unit.HP > 0 && unit.Position.Col == position.Col && unit.Position.Row == position.Row {
				return unit
			}
		}
	}
	return nil
}

func (e *Engine) findBuildingAt(position Position) *models.Building {
	for _, player := range e.State.Players {
		for i := range player.Buildings {
			building := &player.Buildings[i]
			if building.HP > 0 && building.Position.Col == position.Col && building.Position.Row == position.Row {
				return building
			}
		}
	}
	return nil
}

func (e *Engine) isUnlocked(playerID string, requiredLevel int) bool {
	if requiredLevel <= 1 {
		return true
	}
	maxLevel := 1
	for _, building := range e.State.Players[playerID].Buildings {
		if building.Type == models.BuildingFortress && building.HP > 0 && building.Level > maxLevel {
			maxLevel = building.Level
		}
	}
	return maxLevel >= requiredLevel
}

func (e *Engine) castleBonuses(playerID string) (float64, float64) {
	level := 1
	for _, building := range e.State.Players[playerID].Buildings {
		if building.Type == models.BuildingCastle {
			level = building.Level
			break
		}
	}
	if level == 2 {
		return 0.05, 0
	}
	if level >= 3 {
		return 0.10, 0.10
	}
	return 0, 0
}

func (e *Engine) applyCastleBonuses(playerID string) {
	atkBonus, hpBonus := e.castleBonuses(playerID)
	player := e.State.Players[playerID]
	for i := range player.Units {
		unit := &player.Units[i]
		def, ok := unitDefinitions[string(unit.Type)]
		if !ok {
			continue
		}
		unit.ATK = int(float64(def.ATK) * (1 + atkBonus))
		newMax := int(float64(def.HP) * (1 + hpBonus))
		delta := newMax - unit.MaxHP
		unit.MaxHP = newMax
		unit.HP += delta
		if unit.HP > unit.MaxHP {
			unit.HP = unit.MaxHP
		}
	}
}

func (e *Engine) isBuildingOccupied(position Position) bool {
	for _, player := range e.State.Players {
		for _, building := range player.Buildings {
			if building.HP > 0 && building.Position.Col == position.Col && building.Position.Row == position.Row {
				return true
			}
		}
	}
	return false
}

func (e *Engine) isOccupied(position Position) bool {
	for _, player := range e.State.Players {
		for _, building := range player.Buildings {
			if building.HP > 0 && building.Position.Col == position.Col && building.Position.Row == position.Row {
				return true
			}
		}
		for _, unit := range player.Units {
			if unit.HP > 0 && unit.Position.Col == position.Col && unit.Position.Row == position.Row {
				return true
			}
		}
	}
	return false
}

func (e *Engine) findUnit(id string) *models.Unit {
	for _, player := range e.State.Players {
		for i := range player.Units {
			if player.Units[i].ID == id && player.Units[i].HP > 0 {
				return &player.Units[i]
			}
		}
	}
	return nil
}

func (e *Engine) finish(winner, reason string) {
	if e.State.Phase == "finished" {
		return
	}
	e.State.Phase = "finished"
	e.State.Winner = winner
	e.State.EndReason = reason
	e.events = append(e.events, Event{Type: "GAME_OVER", PlayerID: winner, Tick: e.tick, Reason: reason})
}

func (e *Engine) evaluateVictory() {
	castleHP := map[string]int{PlayerOne: 0, PlayerTwo: 0}
	for _, playerID := range []string{PlayerOne, PlayerTwo} {
		for _, building := range e.State.Players[playerID].Buildings {
			if building.Type == models.BuildingCastle {
				castleHP[playerID] = building.HP
			}
		}
	}
	p1Destroyed, p2Destroyed := castleHP[PlayerOne] <= 0, castleHP[PlayerTwo] <= 0
	if p1Destroyed && p2Destroyed {
		e.finish("", "mutual_destruction_draw")
		return
	}
	if p1Destroyed {
		e.finish(PlayerTwo, "castle_destroyed")
		return
	}
	if p2Destroyed {
		e.finish(PlayerOne, "castle_destroyed")
		return
	}
	if e.State.TimeRemaining == 0 {
		castleHP := map[string]int{PlayerOne: 0, PlayerTwo: 0}
		for playerID, player := range e.State.Players {
			for _, building := range player.Buildings {
				if building.Type == models.BuildingCastle {
					castleHP[playerID] = building.HP
				}
			}
		}
		if castleHP[PlayerOne] > castleHP[PlayerTwo] {
			e.finish(PlayerOne, "timeout")
		} else if castleHP[PlayerTwo] > castleHP[PlayerOne] {
			e.finish(PlayerTwo, "timeout")
		} else {
			e.finish("", "timeout_draw")
		}
	}
}

func (e *Engine) playerOrder() []string {
	// Rotate initiative once per simulation second. Attack and fortress
	// cooldowns are expressed in seconds, so alternating by individual tick
	// would repeatedly favor one side on every eligible combat tick.
	if (e.tick/TickRate)%2 == 0 {
		return []string{PlayerOne, PlayerTwo}
	}
	return []string{PlayerTwo, PlayerOne}
}

func (e *Engine) updateEconomy() {
	for _, playerID := range []string{PlayerOne, PlayerTwo} {
		player := e.State.Players[playerID]
		for _, building := range player.Buildings {
			if building.Type != models.BuildingVillage || building.HP <= 0 {
				continue
			}
			production := villageProduction(building.Level)
			player.Resources.Gold += production.Gold
			player.Resources.Wood += production.Wood
			player.Resources.Stone += production.Stone
		}
	}
}

func (e *Engine) movementPathClear(unitType string, from, to Position) bool {
	if unitType == "cavalry" {
		return true // Knight movement jumps over intervening entities.
	}
	dc, dr := to.Col-from.Col, to.Row-from.Row
	steps := max(abs(dc), abs(dr))
	if steps <= 1 {
		return true
	}
	stepCol, stepRow := sign(dc), sign(dr)
	for i := 1; i < steps; i++ {
		intermediate := Position{Col: from.Col + stepCol*i, Row: from.Row + stepRow*i}
		if unitType == "reaper" {
			if e.isBuildingOccupied(intermediate) {
				return false
			}
		} else if e.isOccupied(intermediate) {
			return false
		}
	}
	return true
}

func (e *Engine) updateMovement() {
	if e.tick%5 != 0 {
		return
	}
	type movementIntent struct {
		playerID string
		unitID   string
		from     Position
		to       Position
	}

	intents := make([]movementIntent, 0)
	for _, playerID := range e.playerOrder() {
		player := e.State.Players[playerID]
		for i := range player.Units {
			unit := player.Units[i]
			if unit.HP <= 0 || e.tick < e.nextMoveTick[unit.ID] ||
				e.hasAttackableEnemy(unit) || e.hasAttackableBuildingTarget(unit) {
				continue
			}

			from := Position{Col: unit.Position.Col, Row: unit.Position.Row}
			var next Position
			var ok, manual bool

			switch {
			case unit.TargetID != "":
				manual = true
				target := e.findUnit(unit.TargetID)
				if target != nil {
					next, ok = e.nextPathStepToUnit(unit, *target)
				} else {
					unit.TargetID = ""
					unit.TargetPosition = nil
					manual = false
				}
			case unit.TargetBuildingID != "":
				manual = true
				target := e.findBuilding(unit.TargetBuildingID)
				if target != nil {
					next, ok = e.nextPathStep(unit, *target)
				} else {
					unit.TargetBuildingID = ""
					unit.TargetPosition = nil
					manual = false
				}
			}

			if !manual {
				if target, exists := e.manualTargets[unit.ID]; exists {
					manual = true
					if from == target {
						unit.State = models.UnitIdle
						continue
					}
					next, ok = e.nextPathStepToPosition(unit, target)
				}
			}

			if !manual {
				if enemy := e.movementEnemyTarget(unit); enemy != nil {
					next, ok = e.nextPathStepToUnit(unit, *enemy)
				}
				if !ok {
					target := e.targetBuilding(playerID, laneForRow(unit.Position.Row))
					if target == nil {
						continue
					}
					next, ok = e.nextPathStep(unit, *target)
				}
			}
			if manual && !ok {
				continue
			}
			if ok {
				intents = append(intents, movementIntent{playerID: playerID, unitID: unit.ID, from: from, to: next})
			}
		}
	}

	destinationCounts := make(map[Position]int, len(intents))
	for _, intent := range intents {
		destinationCounts[intent.to]++
	}
	movedUnits := make([]string, 0, len(intents))
	for _, intent := range intents {
		if destinationCounts[intent.to] != 1 {
			continue
		}
		unit := e.findUnit(intent.unitID)
		if unit == nil || (Position{Col: unit.Position.Col, Row: unit.Position.Row}) != intent.from {
			continue
		}
		previous := intent.from
		unit.Position = modelsPosition(intent.to)
		unit.State = models.UnitMoving
		e.movedSinceSpawn[unit.ID] = true
		if unit.Type == models.UnitBeast && manhattan(previous, intent.to) >= 3 {
			e.beastChargeReady[unit.ID] = true
		}
		e.nextMoveTick[unit.ID] = e.tick + 5
		e.events = append(e.events, Event{Type: "UNIT_MOVED", PlayerID: intent.playerID, EntityID: unit.ID, Tick: e.tick})
		movedUnits = append(movedUnits, unit.ID)
	}
	e.resolveCavalryLandings(movedUnits)
}

func (e *Engine) targetBuilding(playerID string, lane int) *models.Building {
	enemy := e.State.Players[opponent(playerID)]
	// Villages sit on the direct approach to every strategic building. They
	// must be cleared first instead of being treated as invisible path blockers.
	laneRow := []int{1, 4, 7}[lane]
	for i := range enemy.Buildings {
		b := &enemy.Buildings[i]
		if b.Type == models.BuildingVillage && b.HP > 0 && b.Position.Row == laneRow {
			return b
		}
	}
	wanted := models.BuildingCastle
	if lane == 0 || lane == 2 {
		wanted = models.BuildingFortress
	}
	for i := range enemy.Buildings {
		b := &enemy.Buildings[i]
		if b.Type != wanted || b.HP <= 0 {
			continue
		}
		if wanted == models.BuildingFortress {
			if lane == 0 && b.Position.Row == 1 || lane == 2 && b.Position.Row == 7 {
				return b
			}
		} else {
			return b
		}
	}
	return nil
}

func (e *Engine) movementEnemyTarget(unit models.Unit) *models.Unit {
	enemy := e.State.Players[opponent(unit.Owner)]
	var best *models.Unit
	bestDistance := int(^uint(0) >> 1)
	for i := range enemy.Units {
		candidate := &enemy.Units[i]
		if candidate.HP <= 0 || laneForRow(candidate.Position.Row) != laneForRow(unit.Position.Row) {
			continue
		}
		if candidate.Type == models.UnitReaper && !e.reaperIsRevealed(*candidate) {
			continue
		}
		distance := (candidate.Position.Col - unit.Position.Col) * direction(unit.Owner)
		if distance <= 0 {
			continue
		}
		if distance < bestDistance || distance == bestDistance && (best == nil || candidate.HP < best.HP || candidate.HP == best.HP && candidate.ID < best.ID) {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func (e *Engine) nextPathStepToUnit(unit, target models.Unit) (Position, bool) {
	start := Position{Col: unit.Position.Col, Row: unit.Position.Row}
	targetPosition := Position{Col: target.Position.Col, Row: target.Position.Row}
	if canAttackUnit(string(unit.Type), start, targetPosition) && (unit.Type != models.UnitArcher || e.hasClearLine(start, targetPosition)) {
		return Position{}, false
	}
	type node struct {
		position Position
		first    Position
	}
	queue := []node{{position: start}}
	visited := map[Position]bool{start: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, delta := range movementDeltas(string(unit.Type), direction(unit.Owner)) {
			next := Position{Col: current.position.Col + delta.Col, Row: current.position.Row + delta.Row}
			if next.Col < 0 || next.Col >= 12 || next.Row < 0 || next.Row >= 10 || laneForRow(next.Row) != laneForRow(start.Row) || visited[next] {
				continue
			}
			if e.isOccupied(next) && next != start {
				continue
			}
			if !e.movementPathClear(string(unit.Type), current.position, next) {
				continue
			}
			first := next
			if current.position != start {
				first = current.first
			}
			if canAttackUnit(string(unit.Type), next, targetPosition) && (unit.Type != models.UnitArcher || e.hasClearLine(next, targetPosition)) {
				return first, true
			}
			visited[next] = true
			queue = append(queue, node{position: next, first: first})
		}
	}
	return Position{}, false
}

func (e *Engine) nextPathStepToPosition(unit models.Unit, target Position) (Position, bool) {
	start := Position{Col: unit.Position.Col, Row: unit.Position.Row}
	if start == target {
		return Position{}, false
	}
	type node struct {
		position Position
		first    Position
	}
	queue := []node{{position: start}}
	visited := map[Position]bool{start: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, delta := range movementDeltas(string(unit.Type), direction(unit.Owner)) {
			next := Position{Col: current.position.Col + delta.Col, Row: current.position.Row + delta.Row}
			if next.Col < 0 || next.Col >= 12 || next.Row < 0 || next.Row >= 10 ||
				laneForRow(next.Row) != laneForRow(start.Row) || visited[next] {
				continue
			}
			if e.isOccupied(next) && next != start {
				continue
			}
			if !e.movementPathClear(string(unit.Type), current.position, next) {
				continue
			}
			first := next
			if current.position != start {
				first = current.first
			}
			if next == target {
				return first, true
			}
			visited[next] = true
			queue = append(queue, node{position: next, first: first})
		}
	}
	return Position{}, false
}

func (e *Engine) nextPathStep(unit models.Unit, target models.Building) (Position, bool) {
	start := Position{Col: unit.Position.Col, Row: unit.Position.Row}
	lane := laneForRow(start.Row)
	if e.canAttackBuildingFrom(string(unit.Type), start, Position{Col: target.Position.Col, Row: target.Position.Row}) {
		return Position{}, false
	}
	type node struct {
		position Position
		first    Position
	}
	queue := []node{{position: start}}
	visited := map[Position]bool{start: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, delta := range movementDeltas(string(unit.Type), direction(unit.Owner)) {
			next := Position{Col: current.position.Col + delta.Col, Row: current.position.Row + delta.Row}
			if next.Col < 0 || next.Col >= 12 || next.Row < 0 || next.Row >= 10 || laneForRow(next.Row) != lane || visited[next] {
				continue
			}
			if e.isOccupied(next) && !(next.Col == start.Col && next.Row == start.Row) {
				continue
			}
			if !e.movementPathClear(string(unit.Type), current.position, next) {
				continue
			}
			first := next
			if current.position != start {
				first = current.first
			}
			if e.canAttackBuildingFrom(string(unit.Type), next, Position{Col: target.Position.Col, Row: target.Position.Row}) {
				return first, true
			}
			visited[next] = true
			queue = append(queue, node{position: next, first: first})
		}
	}
	return Position{}, false
}

func (e *Engine) hasAttackableEnemy(unit models.Unit) bool {
	return e.selectEnemyUnit(unit) != nil
}

func (e *Engine) hasAttackableBuildingTarget(unit models.Unit) bool {
	if unit.TargetBuildingID == "" {
		return false
	}
	target := e.findBuilding(unit.TargetBuildingID)
	if target == nil || target.Owner == unit.Owner {
		return false
	}
	from := Position{Col: unit.Position.Col, Row: unit.Position.Row}
	to := Position{Col: target.Position.Col, Row: target.Position.Row}
	return e.canAttackBuildingFrom(string(unit.Type), from, to)
}

func (e *Engine) removeDeadUnits() {
	for _, playerID := range []string{PlayerOne, PlayerTwo} {
		player := e.State.Players[playerID]
		alive := player.Units[:0]
		for _, unit := range player.Units {
			if unit.HP > 0 {
				alive = append(alive, unit)
			} else {
				e.events = append(e.events, Event{Type: "UNIT_DIED", PlayerID: unit.Owner, EntityID: unit.ID, Tick: e.tick})
				delete(e.movedSinceSpawn, unit.ID)
				delete(e.firstAttackConsumed, unit.ID)
				delete(e.reaperRevealed, unit.ID)
				delete(e.beastChargeReady, unit.ID)
				delete(e.spawnTick, unit.ID)
				delete(e.nextMoveTick, unit.ID)
				delete(e.nextAttackTick, unit.ID)
				delete(e.manualTargets, unit.ID)
			}
		}
		player.Units = alive
	}
	for _, playerID := range []string{PlayerOne, PlayerTwo} {
		player := e.State.Players[playerID]
		for i := range player.Buildings {
			building := &player.Buildings[i]
			if building.HP <= 0 {
				building.HP = 0
				if !e.buildingDestroyed[building.ID] {
					e.buildingDestroyed[building.ID] = true
					e.events = append(e.events, Event{Type: "BUILDING_DESTROYED", PlayerID: building.Owner, EntityID: building.ID, Tick: e.tick})
				}
			}
		}
	}
}

func (e *Engine) drainEvents() []Event {
	out := append([]Event(nil), e.events...)
	e.events = e.events[:0]
	return out
}

func (e *Engine) String() string {
	return fmt.Sprintf("Kingdom Arena tick=%d phase=%s remaining=%ds", e.tick, e.State.Phase, e.State.TimeRemaining)
}