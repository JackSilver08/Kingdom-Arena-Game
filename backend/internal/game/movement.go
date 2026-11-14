package game

func movementDeltas(unitType string, forward int) []Position {
	switch unitType {
	case "swordsman":
		// The unit may advance up to two cells. A one-cell fallback keeps
		// opposing front lines from deadlocking when the two-cell path is blocked.
		return []Position{{Col: 2 * forward}, {Col: forward}}
	case "archer":
		return []Position{{forward, -1}, {forward, 0}, {forward, 1}, {-forward, -1}, {-forward, 0}, {-forward, 1}, {0, -1}, {0, 1}}
	case "cavalry":
		return []Position{{2 * forward, -1}, {2 * forward, 1}, {forward, -2}, {forward, 2}, {-2 * forward, -1}, {-2 * forward, 1}, {-forward, -2}, {-forward, 2}}
	case "catapult":
		return []Position{{forward, 0}, {-forward, 0}, {0, -1}, {0, 1}}
	case "wizard":
		return []Position{{2 * forward, -2}, {2 * forward, 2}, {-2 * forward, -2}, {-2 * forward, 2}}
	case "beast":
		return []Position{{4 * forward, 0}, {-4 * forward, 0}, {0, -4}, {0, 4}}
	case "dragon_knight":
		return []Position{{2 * forward, 0}, {-2 * forward, 0}, {0, -2}, {0, 2}, {2 * forward, -2}, {2 * forward, 2}, {-2 * forward, -2}, {-2 * forward, 2}}
	case "reaper":
		return []Position{{3 * forward, 0}}
	default:
		return nil
	}
}

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func canAttackUnit(unitType string, attacker, target Position) bool {
	dc, dr := abs(attacker.Col-target.Col), abs(attacker.Row-target.Row)
	switch unitType {
	case "swordsman", "reaper":
		return dc+dr == 1
	case "archer":
		return (dc == 0 || dr == 0) && dc+dr <= 3 && dc+dr > 0
	case "cavalry", "beast":
		return max(dc, dr) == 1
	case "catapult":
		return max(dc, dr) >= 3 && max(dc, dr) <= 5
	case "wizard":
		return dc == dr && dc > 0 && dc <= 3
	case "dragon_knight":
		return max(dc, dr) > 0 && max(dc, dr) <= 2
	default:
		return false
	}
}

func canAttackBuilding(unitType string, attacker, target Position) bool {
	dc, dr := abs(attacker.Col-target.Col), abs(attacker.Row-target.Row)
	switch unitType {
	case "swordsman", "reaper":
		return dc+dr == 1
	case "archer":
		return (dc == 0 || dr == 0) && dc+dr <= 3 && dc+dr > 0
	case "cavalry", "beast":
		return max(dc, dr) == 1
	case "catapult":
		return max(dc, dr) >= 3 && max(dc, dr) <= 5
	case "wizard":
		return dc == dr && dc > 0 && dc <= 3
	case "dragon_knight":
		return max(dc, dr) > 0 && max(dc, dr) <= 2
	default:
		return false
	}
}