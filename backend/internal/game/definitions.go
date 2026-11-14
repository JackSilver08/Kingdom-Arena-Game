package game

import "kingdom-arena/pkg/models"

type Position struct{ Col, Row int }
type Cost struct{ Gold, Wood, Stone float64 }
type UnitDefinition struct {
	HP, ATK, TrainSeconds, UnlockLevel, AttackCooldownTicks int
	Cost                                                    Cost
	AOE                                                     bool
}
type Production struct{ Gold, Wood, Stone float64 }

var unitDefinitions = map[string]UnitDefinition{
	"swordsman":     {HP: 80, ATK: 15, TrainSeconds: 3, UnlockLevel: 1, AttackCooldownTicks: 10, Cost: Cost{Gold: 30}},
	"archer":        {HP: 50, ATK: 20, TrainSeconds: 4, UnlockLevel: 1, AttackCooldownTicks: 13, Cost: Cost{Gold: 40, Wood: 10}},
	"cavalry":       {HP: 100, ATK: 25, TrainSeconds: 5, UnlockLevel: 1, AttackCooldownTicks: 9, Cost: Cost{Gold: 60, Wood: 15}},
	"catapult":      {HP: 60, ATK: 40, TrainSeconds: 8, UnlockLevel: 1, AttackCooldownTicks: 34, Cost: Cost{Gold: 80, Wood: 30, Stone: 20}, AOE: true},
	"wizard":        {HP: 40, ATK: 35, TrainSeconds: 7, UnlockLevel: 2, AttackCooldownTicks: 20, Cost: Cost{Gold: 80, Wood: 20, Stone: 20}, AOE: true},
	"beast":         {HP: 200, ATK: 30, TrainSeconds: 9, UnlockLevel: 2, AttackCooldownTicks: 15, Cost: Cost{Gold: 100, Wood: 30, Stone: 30}},
	"dragon_knight": {HP: 150, ATK: 40, TrainSeconds: 12, UnlockLevel: 3, AttackCooldownTicks: 13, Cost: Cost{Gold: 150, Wood: 50, Stone: 50}},
	"reaper":        {HP: 60, ATK: 50, TrainSeconds: 10, UnlockLevel: 3, AttackCooldownTicks: 17, Cost: Cost{Gold: 120, Wood: 40, Stone: 40}},
}

func villageProduction(level int) Production {
	switch level {
	case 2:
		return Production{Gold: 4, Wood: 2, Stone: 2}
	case 3:
		return Production{Gold: 6, Wood: 3, Stone: 3}
	default:
		return Production{Gold: 2, Wood: 1, Stone: 1}
	}
}

func buildingUpgradeCost(buildingType string, nextLevel int) Cost {
	if nextLevel == 2 {
		switch buildingType {
		case "castle":
			return Cost{Gold: 100, Wood: 50, Stone: 50}
		case "fortress":
			return Cost{Gold: 80, Wood: 40, Stone: 40}
		case "village":
			return Cost{Gold: 60, Wood: 30, Stone: 20}
		}
	}
	if nextLevel == 3 {
		switch buildingType {
		case "castle":
			return Cost{Gold: 200, Wood: 100, Stone: 100}
		case "fortress":
			return Cost{Gold: 160, Wood: 80, Stone: 80}
		case "village":
			return Cost{Gold: 120, Wood: 60, Stone: 40}
		}
	}
	return Cost{}
}

func buildingMaxHP(buildingType string, level int) int {
	switch buildingType {
	case "castle":
		return []int{0, 500, 750, 1000}[level]
	case "fortress":
		return []int{0, 300, 500, 700}[level]
	case "village":
		return []int{0, 150, 250, 400}[level]
	default:
		return 0
	}
}

func fortressStats(level int) (int, int) {
	switch level {
	case 2:
		return 3, 15
	case 3:
		return 3, 25
	default:
		return 2, 10
	}
}

func canAfford(resources models.Resources, cost Cost) bool {
	return resources.Gold >= cost.Gold && resources.Wood >= cost.Wood && resources.Stone >= cost.Stone
}

func spend(resources *models.Resources, cost Cost) {
	resources.Gold -= cost.Gold
	resources.Wood -= cost.Wood
	resources.Stone -= cost.Stone
}

func isPlayer(playerID string) bool { return playerID == PlayerOne || playerID == PlayerTwo }
func opponent(playerID string) string {
	if playerID == PlayerOne {
		return PlayerTwo
	}
	return PlayerOne
}
func direction(playerID string) int {
	if playerID == PlayerOne {
		return 1
	}
	return -1
}
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func manhattan(a, b Position) int { return abs(a.Col-b.Col) + abs(a.Row-b.Row) }
func chebyshev(a, b Position) int { return max(abs(a.Col-b.Col), abs(a.Row-b.Row)) }

func validSpawnPosition(playerID string, position Position) bool {
	if position.Row < 0 || position.Row >= 10 {
		return false
	}
	if playerID == PlayerOne {
		return position.Col >= 0 && position.Col <= 4
	}
	return position.Col >= 7 && position.Col <= 11
}

func laneForRow(row int) int {
	if row <= 2 {
		return 0
	}
	if row <= 6 {
		return 1
	}
	return 2
}

func modelsPosition(position Position) models.Position {
	return models.Position{Col: position.Col, Row: position.Row}
}