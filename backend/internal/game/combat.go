package game

import (
	"sort"

	"kingdom-arena/pkg/models"
)

// Combat resolution is server-authoritative. Player attacks are collected
// against one pre-combat state and their damage is applied as a batch so
// simultaneous lethal attacks do not create a deterministic first-player win.
func (e *Engine) updateCombat() {
	type attackIntent struct {
		attackerID string
		targetID   string
		building   bool
	}
	// Select every target from the same pre-combat state. A unit killed during
	// this phase still gets its already-authorized simultaneous attack.
	attacks := make([]attackIntent, 0)
	for _, playerID := range e.playerOrder() {
		for i := range e.State.Players[playerID].Units {
			unit := &e.State.Players[playerID].Units[i]
			if unit.HP <= 0 || e.tick < e.nextAttackTick[unit.ID] {
				continue
			}
			if target := e.selectEnemyUnit(*unit); target != nil {
				attacks = append(attacks, attackIntent{attackerID: unit.ID, targetID: target.ID})
				continue
			}
			building := e.targetBuilding(playerID, laneForRow(unit.Position.Row))
			if building != nil && e.canAttackBuildingFrom(string(unit.Type), Position{Col: unit.Position.Col, Row: unit.Position.Row}, Position{Col: building.Position.Col, Row: building.Position.Row}) {
				attacks = append(attacks, attackIntent{attackerID: unit.ID, targetID: building.ID, building: true})
			}
		}
	}

	hits := make([]damageHit, 0)
	for _, attack := range attacks {
		attacker := e.findUnit(attack.attackerID)
		if attacker == nil || attacker.HP <= 0 {
			continue
		}
		def := unitDefinitions[string(attacker.Type)]
		beastCharge := attacker.Type == models.UnitBeast && e.beastChargeReady[attacker.ID]
		damage := e.outgoingDamage(attacker)
		e.nextAttackTick[attacker.ID] = e.tick + uint64(max(1, def.AttackCooldownTicks))
		attacker.LastAttackTime = int64(e.tick)
		attacker.State = models.UnitAttacking
		center := Position{}
		if attack.building {
			target := e.findBuilding(attack.targetID)
			if target == nil || target.HP <= 0 {
				continue
			}
			center = Position{Col: target.Position.Col, Row: target.Position.Row}
		} else {
			target := e.findUnit(attack.targetID)
			if target == nil || target.HP <= 0 {
				continue
			}
			center = Position{Col: target.Position.Col, Row: target.Position.Row}
		}

		if def.AOE {
			enemyID := opponent(attacker.Owner)
			for i := range e.State.Players[enemyID].Units {
				victim := &e.State.Players[enemyID].Units[i]
				if victim.HP > 0 && abs(victim.Position.Col-center.Col) <= 1 && abs(victim.Position.Row-center.Row) <= 1 {
					hits = append(hits, damageHit{attackerID: attacker.ID, ownerID: attacker.Owner, targetID: victim.ID, damage: damage, eventType: "DAMAGE", eventEntityID: victim.ID, eventTargetID: attacker.ID})
				}
			}
			for i := range e.State.Players[enemyID].Buildings {
				building := &e.State.Players[enemyID].Buildings[i]
				if building.HP <= 0 || abs(building.Position.Col-center.Col) > 1 || abs(building.Position.Row-center.Row) > 1 {
					continue
				}
				buildingDamage := damage
				if attacker.Type == models.UnitCatapult {
					buildingDamage *= 2
				}
				hits = append(hits, damageHit{attackerID: attacker.ID, ownerID: attacker.Owner, targetID: building.ID, damage: buildingDamage, eventType: "DAMAGE", eventEntityID: building.ID, eventTargetID: attacker.ID})
			}
		} else {
			knockback := beastCharge && !attack.building
			hits = append(hits, damageHit{attackerID: attacker.ID, ownerID: attacker.Owner, targetID: attack.targetID, damage: damage, knockback: knockback, eventType: "DAMAGE", eventEntityID: attack.targetID, eventTargetID: attacker.ID})
		}
	}
	e.applyDamageBatch(hits)
}

func (e *Engine) outgoingDamage(attacker *models.Unit) int {
	damage := attacker.ATK
	if attacker.Type == models.UnitSwordsman && e.movedSinceSpawn[attacker.ID] && !e.firstAttackConsumed[attacker.ID] {
		damage += 5
		e.firstAttackConsumed[attacker.ID] = true
	}
	if attacker.Type == models.UnitReaper && !e.firstAttackConsumed[attacker.ID] {
		damage *= 3
		e.firstAttackConsumed[attacker.ID] = true
		e.reaperRevealed[attacker.ID] = true
	}
	if attacker.Type == models.UnitBeast && e.beastChargeReady[attacker.ID] {
		damage += 20
		e.beastChargeReady[attacker.ID] = false
	}
	return damage
}

func (e *Engine) findBuilding(id string) *models.Building {
	for _, player := range e.State.Players {
		for i := range player.Buildings {
			if player.Buildings[i].ID == id && player.Buildings[i].HP > 0 {
				return &player.Buildings[i]
			}
		}
	}
	return nil
}

func (e *Engine) applyDamageBatch(hits []damageHit) {
	type knockbackIntent struct {
		attackerID string
		targetID   string
	}
	unitTotals := make(map[string]int)
	buildingTotals := make(map[string]int)
	knockbacks := make([]knockbackIntent, 0)
	for _, hit := range hits {
		if hit.damage <= 0 {
			continue
		}
		damage := hit.damage
		if target := e.findUnit(hit.targetID); target != nil {
			damage = e.adjustedUnitDamage(target, damage)
			unitTotals[target.ID] += damage
		} else if target := e.findBuilding(hit.targetID); target != nil {
			buildingTotals[target.ID] += damage
		} else {
			continue
		}
		eventType := hit.eventType
		if eventType == "" {
			eventType = "DAMAGE"
		}
		eventEntityID := hit.eventEntityID
		if eventEntityID == "" {
			eventEntityID = hit.targetID
		}
		eventTargetID := hit.eventTargetID
		if eventTargetID == "" {
			eventTargetID = hit.attackerID
		}
		e.events = append(e.events, Event{Type: eventType, PlayerID: hit.ownerID, EntityID: eventEntityID, TargetID: eventTargetID, Tick: e.tick, Value: damage})
		if hit.knockback {
			knockbacks = append(knockbacks, knockbackIntent{attackerID: hit.attackerID, targetID: hit.targetID})
		}
	}
	for targetID, damage := range unitTotals {
		if target := e.findUnit(targetID); target != nil {
			target.HP -= damage
		}
	}
	for targetID, damage := range buildingTotals {
		if target := e.findBuilding(targetID); target != nil {
			target.HP -= damage
		}
	}
	knocked := make(map[string]bool)
	for _, intent := range knockbacks {
		if knocked[intent.targetID] {
			continue
		}
		target := e.findUnit(intent.targetID)
		attacker := e.findUnit(intent.attackerID)
		if target != nil && target.HP > 0 && attacker != nil {
			e.applyKnockback(attacker, target, Position{Col: target.Position.Col, Row: target.Position.Row})
			knocked[intent.targetID] = true
		}
	}
}

func (e *Engine) selectEnemyUnit(attacker models.Unit) *models.Unit {
	if attacker.TargetID != "" {
		if target := e.findUnit(attacker.TargetID); target != nil && target.Owner != attacker.Owner && (target.Type != models.UnitReaper || e.reaperIsRevealed(*target)) {
			from := Position{Col: attacker.Position.Col, Row: attacker.Position.Row}
			to := Position{Col: target.Position.Col, Row: target.Position.Row}
			if canAttackUnit(string(attacker.Type), from, to) && (attacker.Type != models.UnitArcher || e.hasClearLine(from, to)) {
				return target
			}
		}
	}
	var candidates []*models.Unit
	for i := range e.State.Players[opponent(attacker.Owner)].Units {
		target := &e.State.Players[opponent(attacker.Owner)].Units[i]
		if target.HP <= 0 || target.Type == models.UnitReaper && !e.reaperIsRevealed(*target) {
			continue
		}
		from := Position{Col: attacker.Position.Col, Row: attacker.Position.Row}
		to := Position{Col: target.Position.Col, Row: target.Position.Row}
		if canAttackUnit(string(attacker.Type), from, to) && (attacker.Type != models.UnitArcher || e.hasClearLine(from, to)) {
			candidates = append(candidates, target)
		}
	}
	// Stable ordering preserves deployment order as the final tie-breaker.
	// Random UUIDs must never decide combat priority in otherwise mirrored fights.
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].HP < candidates[j].HP
	})
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

func (e *Engine) attackUnit(attacker, target *models.Unit) {
	def := unitDefinitions[string(attacker.Type)]
	damage := attacker.ATK
	if attacker.Type == models.UnitSwordsman && e.movedSinceSpawn[attacker.ID] && !e.firstAttackConsumed[attacker.ID] {
		damage += 5
		e.firstAttackConsumed[attacker.ID] = true
	}
	if attacker.Type == models.UnitReaper && !e.firstAttackConsumed[attacker.ID] {
		damage *= 3
		e.firstAttackConsumed[attacker.ID] = true
		e.reaperRevealed[attacker.ID] = true
	}
	if attacker.Type == models.UnitBeast && e.beastChargeReady[attacker.ID] {
		damage += 20
		e.beastChargeReady[attacker.ID] = false
	}

	if def.AOE {
		center := Position{Col: target.Position.Col, Row: target.Position.Row}
		for i := range e.State.Players[opponent(attacker.Owner)].Units {
			victim := &e.State.Players[opponent(attacker.Owner)].Units[i]
			if victim.HP > 0 && abs(victim.Position.Col-center.Col) <= 1 && abs(victim.Position.Row-center.Row) <= 1 {
				applied := e.applyUnitDamage(victim, damage)
				e.emitDamage(attacker, victim.ID, applied)
			}
		}
		for i := range e.State.Players[opponent(attacker.Owner)].Buildings {
			building := &e.State.Players[opponent(attacker.Owner)].Buildings[i]
			if building.HP > 0 && abs(building.Position.Col-center.Col) <= 1 && abs(building.Position.Row-center.Row) <= 1 {
				buildingDamage := damage
				if attacker.Type == models.UnitCatapult {
					buildingDamage *= 2
				}
				building.HP -= buildingDamage
				e.emitDamage(attacker, building.ID, buildingDamage)
			}
		}
	} else {
		oldPosition := Position{Col: target.Position.Col, Row: target.Position.Row}
		chargeHit := attacker.Type == models.UnitBeast && damage > attacker.ATK
		applied := e.applyUnitDamage(target, damage)
		if chargeHit && target.HP > 0 {
			e.applyKnockback(attacker, target, oldPosition)
		}
		e.emitDamage(attacker, target.ID, applied)
	}
	if attacker.Type == models.UnitReaper {
		e.reaperRevealed[attacker.ID] = true
	}
	e.nextAttackTick[attacker.ID] = e.tick + uint64(max(1, def.AttackCooldownTicks))
	attacker.LastAttackTime = int64(e.tick)
	attacker.State = models.UnitAttacking
}

func (e *Engine) attackBuilding(attacker *models.Unit, target *models.Building) {
	damage := attacker.ATK
	if attacker.Type == models.UnitBeast && e.beastChargeReady[attacker.ID] {
		damage += 20
		e.beastChargeReady[attacker.ID] = false
	}
	if attacker.Type == models.UnitReaper && !e.firstAttackConsumed[attacker.ID] {
		damage *= 3
		e.firstAttackConsumed[attacker.ID] = true
		e.reaperRevealed[attacker.ID] = true
	}
	if attacker.Type == models.UnitReaper {
		e.reaperRevealed[attacker.ID] = true
	}

	def := unitDefinitions[string(attacker.Type)]
	if def.AOE {
		center := Position{Col: target.Position.Col, Row: target.Position.Row}
		enemyID := opponent(attacker.Owner)
		for i := range e.State.Players[enemyID].Units {
			victim := &e.State.Players[enemyID].Units[i]
			if victim.HP > 0 && abs(victim.Position.Col-center.Col) <= 1 && abs(victim.Position.Row-center.Row) <= 1 {
				applied := e.applyUnitDamage(victim, damage)
				e.emitDamage(attacker, victim.ID, applied)
			}
		}
		for i := range e.State.Players[enemyID].Buildings {
			building := &e.State.Players[enemyID].Buildings[i]
			if building.HP <= 0 || abs(building.Position.Col-center.Col) > 1 || abs(building.Position.Row-center.Row) > 1 {
				continue
			}
			buildingDamage := damage
			if attacker.Type == models.UnitCatapult {
				buildingDamage *= 2
			}
			building.HP -= buildingDamage
			e.emitDamage(attacker, building.ID, buildingDamage)
		}
	} else {
		target.HP -= damage
		e.emitDamage(attacker, target.ID, damage)
	}
	e.nextAttackTick[attacker.ID] = e.tick + uint64(max(1, def.AttackCooldownTicks))
	attacker.LastAttackTime = int64(e.tick)
	attacker.State = models.UnitAttacking
}

func (e *Engine) emitDamage(attacker *models.Unit, targetID string, damage int) {
	e.events = append(e.events, Event{Type: "DAMAGE", PlayerID: attacker.Owner, EntityID: targetID, TargetID: attacker.ID, Tick: e.tick, Value: damage})
}

func (e *Engine) hasClearLine(from, to Position) bool {
	if from.Col != to.Col && from.Row != to.Row {
		return false
	}
	steps := max(abs(to.Col-from.Col), abs(to.Row-from.Row))
	if steps <= 1 {
		return true
	}
	stepCol, stepRow := sign(to.Col-from.Col), sign(to.Row-from.Row)
	for i := 1; i < steps; i++ {
		if e.isOccupied(Position{Col: from.Col + stepCol*i, Row: from.Row + stepRow*i}) {
			return false
		}
	}
	return true
}

func (e *Engine) canAttackBuildingFrom(unitType string, from, to Position) bool {
	if !canAttackBuilding(unitType, from, to) {
		return false
	}
	if unitType == "archer" {
		return e.hasClearLine(from, to)
	}
	return true
}

func (e *Engine) reaperIsRevealed(unit models.Unit) bool {
	if e.reaperRevealed[unit.ID] {
		return true
	}
	return e.tick >= e.spawnTick[unit.ID]+3*TickRate
}

func (e *Engine) adjustedUnitDamage(target *models.Unit, damage int) int {
	if target.Type == models.UnitReaper && e.reaperIsRevealed(*target) {
		damage = (damage*120 + 99) / 100
	}
	return damage
}

func (e *Engine) applyUnitDamage(target *models.Unit, damage int) int {
	damage = e.adjustedUnitDamage(target, damage)
	target.HP -= damage
	return damage
}

func (e *Engine) applyKnockback(attacker, target *models.Unit, oldPosition Position) {
	if target.Type == models.UnitDragonKnight {
		return
	}
	delta := Position{Col: sign(oldPosition.Col - attacker.Position.Col), Row: sign(oldPosition.Row - attacker.Position.Row)}
	destination := Position{Col: oldPosition.Col + delta.Col, Row: oldPosition.Row + delta.Row}
	if destination.Col < 0 || destination.Col >= 12 || destination.Row < 0 || destination.Row >= 10 || e.isOccupied(destination) {
		return
	}
	target.Position = models.Position{Col: destination.Col, Row: destination.Row}
	e.events = append(e.events, Event{Type: "UNIT_KNOCKBACK", PlayerID: attacker.Owner, EntityID: target.ID, Tick: e.tick})
}

func (e *Engine) resolveCavalryLandings(movedUnitIDs []string) {
	hits := make([]damageHit, 0)
	for _, unitID := range movedUnitIDs {
		cavalry := e.findUnit(unitID)
		if cavalry == nil || cavalry.Type != models.UnitCavalry {
			continue
		}
		var target *models.Unit
		for i := range e.State.Players[opponent(cavalry.Owner)].Units {
			candidate := &e.State.Players[opponent(cavalry.Owner)].Units[i]
			if candidate.HP <= 0 || chebyshev(Position{Col: cavalry.Position.Col, Row: cavalry.Position.Row}, Position{Col: candidate.Position.Col, Row: candidate.Position.Row}) != 1 {
				continue
			}
			if target == nil || candidate.HP < target.HP {
				target = candidate
			}
		}
		if target == nil {
			continue
		}
		hits = append(hits, damageHit{
			attackerID: cavalry.ID, ownerID: cavalry.Owner, targetID: target.ID,
			damage: cavalry.ATK, eventType: "CAVALRY_IMPACT",
			eventEntityID: cavalry.ID, eventTargetID: target.ID,
		})
		e.nextAttackTick[cavalry.ID] = e.tick + uint64(unitDefinitions[string(cavalry.Type)].AttackCooldownTicks)
	}
	e.applyDamageBatch(hits)
}

func (e *Engine) resolveCavalryLanding(cavalry *models.Unit) {
	if cavalry != nil {
		e.resolveCavalryLandings([]string{cavalry.ID})
	}
}

func (e *Engine) updateFortresses() {
	hits := make([]damageHit, 0)
	for _, playerID := range e.playerOrder() {
		player := e.State.Players[playerID]
		for i := range player.Buildings {
			building := &player.Buildings[i]
			if building.Type != models.BuildingFortress || building.HP <= 0 || e.tick < e.nextFortressAttack[building.ID] {
				continue
			}
			rangeCells, damage := fortressStats(building.Level)
			var target *models.Unit
			for j := range e.State.Players[opponent(playerID)].Units {
				candidate := &e.State.Players[opponent(playerID)].Units[j]
				if candidate.HP <= 0 || candidate.Type == models.UnitReaper && !e.reaperIsRevealed(*candidate) || manhattan(Position{Col: building.Position.Col, Row: building.Position.Row}, Position{Col: candidate.Position.Col, Row: candidate.Position.Row}) > rangeCells {
					continue
				}
				if target == nil || candidate.HP < target.HP {
					target = candidate
				}
			}
			if target != nil {
				hits = append(hits, damageHit{
					ownerID: playerID, targetID: target.ID, damage: damage,
					eventType: "FORTRESS_HIT", eventEntityID: building.ID, eventTargetID: target.ID,
				})
				e.nextFortressAttack[building.ID] = e.tick + TickRate
			}
		}
	}
	e.applyDamageBatch(hits)
}

func (e *Engine) updateDragonBreaths() {
	hits := make([]damageHit, 0)
	for _, playerID := range e.playerOrder() {
		for i := range e.State.Players[playerID].Units {
			dragon := &e.State.Players[playerID].Units[i]
			if dragon.HP <= 0 || dragon.Type != models.UnitDragonKnight || e.tick < e.lastDragonBreathTick[dragon.ID]+100 {
				continue
			}
			for j := range e.State.Players[opponent(playerID)].Units {
				target := &e.State.Players[opponent(playerID)].Units[j]
				forwardDistance := (target.Position.Col - dragon.Position.Col) * direction(playerID)
				if target.HP > 0 && target.Position.Row == dragon.Position.Row && forwardDistance >= 1 && forwardDistance <= 3 {
					hits = append(hits, damageHit{
						attackerID: dragon.ID, ownerID: playerID, targetID: target.ID, damage: 20,
						eventType: "DRAGON_BREATH", eventEntityID: dragon.ID, eventTargetID: target.ID,
					})
				}
			}
			e.lastDragonBreathTick[dragon.ID] = e.tick
		}
	}
	e.applyDamageBatch(hits)
}