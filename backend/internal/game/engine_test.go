package game

import (
	"errors"
	"testing"

	"kingdom-arena/pkg/models"
)

func newTestEngine() *Engine { return NewEngine(models.NewGameState()) }

func TestDeployValidatesResourcesAndSpawnZone(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatalf("valid deploy failed: %v", err)
	}
	if got := e.State.Players[PlayerOne].Resources.Gold; got != 20 {
		t.Fatalf("gold = %v, want 20", got)
	}
	if len(e.State.Players[PlayerOne].Units) != 1 {
		t.Fatalf("unit count = %d, want 1", len(e.State.Players[PlayerOne].Units))
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 7, Row: 4}); !errors.Is(err, ErrInvalidPosition) {
		t.Fatalf("cross-side deploy error = %v", err)
	}
}

func TestDeployRejectsOccupiedAndLockedUnits(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "archer", Col: 4, Row: 4}); !errors.Is(err, ErrOccupiedPosition) {
		t.Fatalf("occupied error = %v", err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "wizard", Col: 4, Row: 3}); !errors.Is(err, ErrUnitLocked) {
		t.Fatalf("locked error = %v", err)
	}
}

func TestCommandIDIsIdempotent(t *testing.T) {
	e := newTestEngine()
	cmd := Command{Type: "DEPLOY_UNIT", CommandID: "same-id", UnitType: "swordsman", Col: 4, Row: 4}
	if err := e.ApplyCommand(PlayerOne, cmd); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerOne, cmd); !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("duplicate error = %v", err)
	}
	if got := e.State.Players[PlayerOne].Resources.Gold; got != 20 {
		t.Fatalf("gold = %v, want 20", got)
	}
}

func TestEconomyProducesEverySecond(t *testing.T) {
	e := newTestEngine()
	for i := 0; i < TickRate; i++ {
		e.Tick()
	}
	resources := e.State.Players[PlayerOne].Resources
	if resources.Gold != 56 || resources.Wood != 23 || resources.Stone != 23 {
		t.Fatalf("unexpected resources after one second: %+v", resources)
	}
}

func TestUpgradeConsumesResourcesAndIncreasesBuildingHP(t *testing.T) {
	e := newTestEngine()
	village := &e.State.Players[PlayerOne].Buildings[3]
	e.State.Players[PlayerOne].Resources = models.Resources{Gold: 60, Wood: 30, Stone: 20}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "UPGRADE", BuildingID: village.ID}); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if village.Level != 2 || village.MaxHP != 250 || village.HP != 250 {
		t.Fatalf("unexpected village after upgrade: %+v", village)
	}
	if got := e.State.Players[PlayerOne].Resources.Gold; got != 0 {
		t.Fatalf("gold = %v, want 0", got)
	}
}

func TestUnitUnlocksAtFortressLevelTwo(t *testing.T) {
	e := newTestEngine()
	fortress := e.State.Players[PlayerOne].Buildings[1]
	// Provide the exact resources needed for the upgrade.
	e.State.Players[PlayerOne].Resources = models.Resources{Gold: 160, Wood: 60, Stone: 60}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "UPGRADE", BuildingID: fortress.ID}); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "wizard", Col: 4, Row: 4}); err != nil {
		t.Fatalf("wizard should unlock at fortress level 2: %v", err)
	}
}

func TestSurrenderFinishesMatch(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "SURRENDER"}); err != nil {
		t.Fatal(err)
	}
	if e.State.Phase != "finished" || e.State.Winner != PlayerTwo {
		t.Fatalf("unexpected result: phase=%s winner=%s", e.State.Phase, e.State.Winner)
	}
	if err := e.ApplyCommand(PlayerTwo, Command{Type: "SURRENDER"}); !errors.Is(err, ErrMatchFinished) {
		t.Fatalf("post-match command error = %v", err)
	}
}

func TestTimeoutUsesCastleHPAndCanDraw(t *testing.T) {
	e := newTestEngine()
	e.State.TimeRemaining = 1
	if e.State.Players[PlayerOne].Buildings[0].HP == 0 {
		t.Fatal("invalid fixture")
	}
	for i := 0; i < TickRate; i++ {
		e.Tick()
	}
	if e.State.Phase != "finished" || e.State.Winner != "" {
		t.Fatalf("equal castle HP should draw: %+v", e.State)
	}
}

func TestMovementProfilesAreDistinct(t *testing.T) {
	cases := map[string]Position{
		"swordsman": {Col: 2}, "archer": {Col: 1, Row: 1}, "cavalry": {Col: 2, Row: 1},
		"catapult": {Col: 1}, "wizard": {Col: 2, Row: 2}, "beast": {Col: 4},
		"dragon_knight": {Col: 2}, "reaper": {Col: 3},
	}
	for unitType, expected := range cases {
		found := false
		for _, delta := range movementDeltas(unitType, 1) {
			if delta == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s movement profile does not include %+v", unitType, expected)
		}
	}
}

func TestBotCanDeployAndSnapshotIsIsolated(t *testing.T) {
	e := newTestEngine()
	if err := e.RunBotTurn(PlayerOne); err != nil {
		t.Fatalf("bot turn failed: %v", err)
	}
	if len(e.State.Players[PlayerOne].Units) == 0 {
		t.Fatal("bot did not deploy a unit")
	}
	snapshot, err := e.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Players[PlayerOne].Resources.Gold = 999
	if e.State.Players[PlayerOne].Resources.Gold == 999 {
		t.Fatal("snapshot must not alias authoritative state")
	}
}

func TestCatapultBuildingDamageMultiplier(t *testing.T) {
	e := newTestEngine()
	attacker := models.Unit{ID: "siege", Type: models.UnitCatapult, Owner: PlayerOne, Position: models.Position{Col: 6, Row: 4}, HP: 60, MaxHP: 60, ATK: 40}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, attacker)
	target := &e.State.Players[PlayerTwo].Buildings[0]
	target.Position = models.Position{Col: 11, Row: 4}
	before := target.HP
	attackerPtr := &e.State.Players[PlayerOne].Units[0]
	e.attackBuilding(attackerPtr, target)
	if got := before - target.HP; got != 80 {
		t.Fatalf("catapult building damage = %d, want 80", got)
	}
}

func TestCastleDestructionEndsMatch(t *testing.T) {
	e := newTestEngine()
	target := &e.State.Players[PlayerTwo].Buildings[0]
	target.HP = 40
	e.State.Players[PlayerTwo].Buildings[4].HP = 0 // The mid-lane Village has already been cleared.
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, models.Unit{
		ID: "siege-castle", Type: models.UnitCatapult, Owner: PlayerOne,
		Position: models.Position{Col: 6, Row: 4}, HP: 60, MaxHP: 60, ATK: 40,
	})
	e.Tick()
	if e.State.Phase != "finished" || e.State.Winner != PlayerOne || e.State.EndReason != "castle_destroyed" {
		t.Fatalf("unexpected castle destruction result: phase=%s winner=%s reason=%s", e.State.Phase, e.State.Winner, e.State.EndReason)
	}
}

func TestReaperIsHiddenForFirstThreeSeconds(t *testing.T) {
	e := newTestEngine()
	reaper := models.Unit{ID: "reaper-hidden", Type: models.UnitReaper, Owner: PlayerTwo, Position: models.Position{Col: 5, Row: 4}, HP: 60, MaxHP: 60, ATK: 50}
	swordsman := models.Unit{ID: "sword-hunter", Type: models.UnitSwordsman, Owner: PlayerOne, Position: models.Position{Col: 4, Row: 4}, HP: 80, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, reaper)
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, swordsman)
	e.spawnTick[reaper.ID] = 0
	if got := e.selectEnemyUnit(swordsman); got != nil {
		t.Fatal("reaper should be hidden during its first three seconds")
	}
	e.tick = 30
	if got := e.selectEnemyUnit(swordsman); got == nil || got.ID != reaper.ID {
		t.Fatal("reaper should be targetable after three seconds")
	}
}

func TestBeastChargeAddsDamageAndKnocksBack(t *testing.T) {
	e := newTestEngine()
	beast := models.Unit{ID: "beast-charge", Type: models.UnitBeast, Owner: PlayerOne, Position: models.Position{Col: 5, Row: 4}, HP: 200, MaxHP: 200, ATK: 30}
	enemy := models.Unit{ID: "enemy-unit", Type: models.UnitSwordsman, Owner: PlayerTwo, Position: models.Position{Col: 6, Row: 4}, HP: 100, MaxHP: 100, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, beast)
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, enemy)
	e.beastChargeReady[beast.ID] = true
	e.attackUnit(&e.State.Players[PlayerOne].Units[0], &e.State.Players[PlayerTwo].Units[0])
	if got := e.State.Players[PlayerTwo].Units[0].HP; got != 50 {
		t.Fatalf("enemy HP = %d, want 50", got)
	}
	if got := e.State.Players[PlayerTwo].Units[0].Position.Col; got != 7 {
		t.Fatalf("enemy knockback col = %d, want 7", got)
	}
}

func TestArcherCannotShootThroughOccupiedCell(t *testing.T) {
	e := newTestEngine()
	archer := models.Unit{ID: "archer", Type: models.UnitArcher, Owner: PlayerOne, Position: models.Position{Col: 4, Row: 4}, HP: 50, MaxHP: 50, ATK: 20}
	blocker := models.Unit{ID: "blocker", Type: models.UnitSwordsman, Owner: PlayerOne, Position: models.Position{Col: 5, Row: 4}, HP: 80, MaxHP: 80, ATK: 15}
	enemy := models.Unit{ID: "enemy", Type: models.UnitSwordsman, Owner: PlayerTwo, Position: models.Position{Col: 6, Row: 4}, HP: 80, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, archer, blocker)
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, enemy)
	if got := e.selectEnemyUnit(archer); got != nil {
		t.Fatal("archer must not target through an occupied cell")
	}
}

func TestCavalryLandingDealsImpactDamage(t *testing.T) {
	e := newTestEngine()
	cavalry := models.Unit{ID: "cavalry", Type: models.UnitCavalry, Owner: PlayerOne, Position: models.Position{Col: 4, Row: 4}, HP: 100, MaxHP: 100, ATK: 25}
	enemy := models.Unit{ID: "enemy", Type: models.UnitSwordsman, Owner: PlayerTwo, Position: models.Position{Col: 7, Row: 4}, HP: 80, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, cavalry)
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, enemy)
	for i := 0; i < 5; i++ {
		e.Tick()
	}
	if e.State.Players[PlayerTwo].Units[0].HP >= 80 {
		t.Fatal("cavalry landing should deal immediate impact damage")
	}
}

func TestDestroyedFortressCannotUnlockAdvancedUnits(t *testing.T) {
	e := newTestEngine()
	fortress := &e.State.Players[PlayerOne].Buildings[1]
	fortress.Level = 2
	fortress.HP = 0
	if e.isUnlocked(PlayerOne, 2) {
		t.Fatal("destroyed fortress must not unlock advanced units")
	}
}

func TestSwordsmanAdvancesTowardMidLaneCastle(t *testing.T) {
	e := newTestEngine()
	unit := models.Unit{ID: "marching-sword", Type: models.UnitSwordsman, Owner: PlayerOne, Position: models.Position{Col: 4, Row: 4}, HP: 80, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, unit)
	target := e.targetBuilding(PlayerOne, 1)
	if target == nil {
		t.Fatal("missing target building")
	}
	if step, ok := e.nextPathStep(unit, *target); !ok {
		t.Fatalf("no path step: %+v", target)
	} else {
		t.Logf("next path step: %+v", step)
	}
	for i := 0; i < 5; i++ {
		e.Tick()
	}
	if got := e.State.Players[PlayerOne].Units[0].Position.Col; got != 6 {
		t.Fatalf("swordsman column after one movement interval = %d, want 6", got)
	}
}

func TestBotVersusBotSimulationRemainsSymmetric(t *testing.T) {
	e := newTestEngine()
	for second := 0; second < 120 && e.State.Phase == "playing"; second++ {
		if err := e.RunBotTurn(PlayerOne); err != nil {
			t.Fatalf("p1 bot: %v", err)
		}
		if err := e.RunBotTurn(PlayerTwo); err != nil {
			t.Fatalf("p2 bot: %v", err)
		}
		for tick := 0; tick < TickRate; tick++ {
			e.Tick()
		}
	}
	p1Castle, p2Castle := 0, 0
	for _, b := range e.State.Players[PlayerOne].Buildings {
		if b.Type == models.BuildingCastle {
			p1Castle = b.HP
		}
	}
	for _, b := range e.State.Players[PlayerTwo].Buildings {
		if b.Type == models.BuildingCastle {
			p2Castle = b.HP
		}
	}
	if p1Castle != p2Castle {
		t.Fatalf("mirrored bot simulation is asymmetric: p1 castle=%d p2 castle=%d", p1Castle, p2Castle)
	}
	for i := range e.State.Players[PlayerOne].Buildings {
		p1Building := e.State.Players[PlayerOne].Buildings[i]
		p2Building := e.State.Players[PlayerTwo].Buildings[i]
		if p1Building.Type != p2Building.Type || p1Building.HP != p2Building.HP || p1Building.Level != p2Building.Level {
			t.Fatalf("mirrored building state diverged at index %d: p1=%+v p2=%+v", i, p1Building, p2Building)
		}
	}
}

func TestSnapshotForHidesUnrevealedEnemyReaper(t *testing.T) {
	e := newTestEngine()
	reaper := models.Unit{ID: "reaper-hidden", Type: models.UnitReaper, Owner: PlayerTwo, Position: models.Position{Col: 6, Row: 4}, HP: 60, MaxHP: 60, ATK: 50}
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, reaper)
	e.spawnTick[reaper.ID] = 0
	view, err := e.SnapshotFor(PlayerOne)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Players[PlayerTwo].Units) != 0 {
		t.Fatal("hidden Reaper leaked into enemy player snapshot")
	}
	e.tick = 30
	view, err = e.SnapshotFor(PlayerOne)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Players[PlayerTwo].Units) != 1 {
		t.Fatal("Reaper should appear after three seconds")
	}
}

func TestMutualCastleDestructionIsDraw(t *testing.T) {
	e := newTestEngine()
	e.State.Players[PlayerOne].Buildings[0].HP = 0
	e.State.Players[PlayerTwo].Buildings[0].HP = 0
	e.evaluateVictory()
	if e.State.Phase != "finished" || e.State.Winner != "" || e.State.EndReason != "mutual_destruction_draw" {
		t.Fatalf("unexpected mutual destruction result: %+v", e.State)
	}
}

func TestDeploymentCooldownIsEnforced(t *testing.T) {
	e := newTestEngine()
	first := Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}
	if err := e.ApplyCommand(PlayerOne, first); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 3, Row: 4}); !errors.Is(err, ErrUnitCooldown) {
		t.Fatalf("expected deployment cooldown, got %v", err)
	}
}

func TestRevealedReaperTakesTwentyPercentAdditionalDamage(t *testing.T) {
	e := newTestEngine()
	reaper := models.Unit{ID: "reaper-revealed", Type: models.UnitReaper, Owner: PlayerTwo, Position: models.Position{Col: 6, Row: 4}, HP: 60, MaxHP: 60, ATK: 50}
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, reaper)
	e.reaperRevealed[reaper.ID] = true
	applied := e.applyUnitDamage(&e.State.Players[PlayerTwo].Units[0], 10)
	if applied != 12 || e.State.Players[PlayerTwo].Units[0].HP != 48 {
		t.Fatalf("revealed Reaper damage=%d hp=%d, want damage 12 hp 48", applied, e.State.Players[PlayerTwo].Units[0].HP)
	}
}

func TestCatapultBuildingAttackAlsoDamagesNearbyEntities(t *testing.T) {
	e := newTestEngine()
	e.State.Players[PlayerTwo].Buildings[4].HP = 100
	nearby := models.Unit{ID: "nearby-enemy", Type: models.UnitSwordsman, Owner: PlayerTwo, Position: models.Position{Col: 9, Row: 5}, HP: 80, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, nearby)
	catapult := models.Unit{ID: "catapult", Type: models.UnitCatapult, Owner: PlayerOne, Position: models.Position{Col: 6, Row: 4}, HP: 60, MaxHP: 60, ATK: 40}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, catapult)
	e.attackBuilding(&e.State.Players[PlayerOne].Units[0], &e.State.Players[PlayerTwo].Buildings[4])
	if got := e.State.Players[PlayerTwo].Buildings[4].HP; got != 20 {
		t.Fatalf("target Village HP=%d, want 20", got)
	}
	if got := e.State.Players[PlayerTwo].Units[0].HP; got != 40 {
		t.Fatalf("nearby unit HP=%d, want 40", got)
	}
}

func TestSwordsmanMovementIncludesOneAndTwoCellAdvance(t *testing.T) {
	deltas := movementDeltas("swordsman", 1)
	wantOne := Position{Col: 1, Row: 0}
	wantTwo := Position{Col: 2, Row: 0}
	foundOne, foundTwo := false, false
	for _, delta := range deltas {
		foundOne = foundOne || delta == wantOne
		foundTwo = foundTwo || delta == wantTwo
	}
	if !foundOne || !foundTwo {
		t.Fatalf("swordsman deltas=%+v, want one- and two-cell forward options", deltas)
	}
}

func TestOpposingSwordsmenEngageInsteadOfDeadlocking(t *testing.T) {
	e := newTestEngine()
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, models.Unit{
		ID: "p1-front", Type: models.UnitSwordsman, Owner: PlayerOne,
		Position: models.Position{Col: 4, Row: 4}, HP: 80, MaxHP: 80, ATK: 15,
	})
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, models.Unit{
		ID: "p2-front", Type: models.UnitSwordsman, Owner: PlayerTwo,
		Position: models.Position{Col: 7, Row: 4}, HP: 80, MaxHP: 80, ATK: 15,
	})
	for i := 0; i < 5; i++ {
		e.Tick()
	}
	p1 := e.State.Players[PlayerOne].Units[0]
	p2 := e.State.Players[PlayerTwo].Units[0]
	if p1.HP >= 80 || p2.HP >= 80 {
		t.Fatalf("front-line units failed to engage: p1HP=%d p2HP=%d", p1.HP, p2.HP)
	}
	if chebyshev(Position{p1.Position.Col, p1.Position.Row}, Position{p2.Position.Col, p2.Position.Row}) > 1 {
		t.Fatalf("front-line units did not reach combat range: p1=%+v p2=%+v", p1.Position, p2.Position)
	}
}

func TestBeastChargeThroughAuthoritativeCombatAddsDamageAndKnockback(t *testing.T) {
	e := newTestEngine()
	beast := models.Unit{ID: "beast-batch", Type: models.UnitBeast, Owner: PlayerOne, Position: models.Position{Col: 5, Row: 4}, HP: 200, MaxHP: 200, ATK: 30}
	enemy := models.Unit{ID: "enemy-batch", Type: models.UnitSwordsman, Owner: PlayerTwo, Position: models.Position{Col: 6, Row: 4}, HP: 100, MaxHP: 100, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, beast)
	e.State.Players[PlayerTwo].Units = append(e.State.Players[PlayerTwo].Units, enemy)
	e.beastChargeReady[beast.ID] = true
	e.updateCombat()
	if got := e.State.Players[PlayerTwo].Units[0].HP; got != 50 {
		t.Fatalf("enemy HP=%d, want 50 after 30+20 charge damage", got)
	}
	if got := e.State.Players[PlayerTwo].Units[0].Position.Col; got != 7 {
		t.Fatalf("enemy knockback col=%d, want 7", got)
	}
}

func TestCastleLevelThreeBuffAppliesToExistingUnits(t *testing.T) {
	e := newTestEngine()
	unit := models.Unit{ID: "buff-target", Type: models.UnitSwordsman, Owner: PlayerOne, Position: models.Position{Col: 4, Row: 4}, HP: 60, MaxHP: 80, ATK: 15}
	e.State.Players[PlayerOne].Units = append(e.State.Players[PlayerOne].Units, unit)
	e.State.Players[PlayerOne].Resources = models.Resources{Gold: 300, Wood: 150, Stone: 150}
	castle := e.State.Players[PlayerOne].Buildings[0]
	if err := e.ApplyCommand(PlayerOne, Command{Type: "UPGRADE", BuildingID: castle.ID}); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "UPGRADE", BuildingID: castle.ID}); err != nil {
		t.Fatal(err)
	}
	upgraded := e.State.Players[PlayerOne].Units[0]
	if upgraded.ATK != 16 || upgraded.MaxHP != 88 || upgraded.HP != 68 {
		t.Fatalf("unexpected buffed unit: atk=%d hp=%d/%d", upgraded.ATK, upgraded.HP, upgraded.MaxHP)
	}
}

func TestMoveToEmptyCellMovesUnitAndStopsAtDestination(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatal(err)
	}
	unitID := e.State.Players[PlayerOne].Units[0].ID
	if err := e.ApplyCommand(PlayerOne, Command{Type: "MOVE_TO", UnitID: unitID, Col: 8, Row: 4}); err != nil {
		t.Fatalf("move order failed: %v", err)
	}
	for i := 0; i < 20; i++ {
		e.Tick()
	}
	unit := e.findUnit(unitID)
	if unit == nil || unit.Position.Col != 8 || unit.Position.Row != 4 {
		t.Fatalf("unit did not reach destination (8,4): %+v", unit)
	}
	if unit.TargetPosition == nil || unit.TargetPosition.Col != 8 {
		t.Fatalf("move order should remain as a hold position: %+v", unit.TargetPosition)
	}
}

func TestMoveToEnemyUnitPrioritizesRightClickedTarget(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatal(err)
	}
	if err := e.ApplyCommand(PlayerTwo, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 7, Row: 4}); err != nil {
		t.Fatal(err)
	}
	attackerID := e.State.Players[PlayerOne].Units[0].ID
	targetID := e.State.Players[PlayerTwo].Units[0].ID
	if err := e.ApplyCommand(PlayerOne, Command{Type: "MOVE_TO", UnitID: attackerID, Col: 7, Row: 4}); err != nil {
		t.Fatalf("attack order failed: %v", err)
	}
	for i := 0; i < 10; i++ {
		e.Tick()
	}
	target := e.findUnit(targetID)
	if target != nil && target.HP >= target.MaxHP {
		t.Fatal("right-clicked enemy did not take damage")
	}
}

func TestMoveToEnemyBuildingMovesIntoAttackRangeAndAttacks(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatal(err)
	}
	unitID := e.State.Players[PlayerOne].Units[0].ID
	building := &e.State.Players[PlayerTwo].Buildings[4]
	before := building.HP
	if err := e.ApplyCommand(PlayerOne, Command{Type: "MOVE_TO", UnitID: unitID, Col: building.Position.Col, Row: building.Position.Row}); err != nil {
		t.Fatalf("building attack order failed: %v", err)
	}
	for i := 0; i < 40; i++ {
		e.Tick()
	}
	if building.HP >= before {
		t.Fatalf("right-clicked building was not attacked: hp=%d", building.HP)
	}
}

func TestMoveToRejectsFriendlyTargetsAndOtherLanes(t *testing.T) {
	e := newTestEngine()
	if err := e.ApplyCommand(PlayerOne, Command{Type: "DEPLOY_UNIT", UnitType: "swordsman", Col: 4, Row: 4}); err != nil {
		t.Fatal(err)
	}
	unitID := e.State.Players[PlayerOne].Units[0].ID
	if err := e.ApplyCommand(PlayerOne, Command{Type: "MOVE_TO", UnitID: unitID, Col: 2, Row: 4}); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("friendly building target error = %v", err)
	}
	if err := e.ApplyCommand(PlayerOne, Command{Type: "MOVE_TO", UnitID: unitID, Col: 8, Row: 1}); !errors.Is(err, ErrInvalidPosition) {
		t.Fatalf("cross-lane target error = %v", err)
	}
}
