package main

import (
	"flag"
	"fmt"
	"log"

	"kingdom-arena/internal/game"
)

func main() {
	seconds := flag.Int("seconds", 60, "simulation duration in seconds (maximum 600)")
	flag.Parse()
	if *seconds < 1 || *seconds > 600 {
		log.Fatal("seconds must be between 1 and 600")
	}

	engine := game.NewEngineFactory().CreateEngine()
	for second := 0; second < *seconds && engine.State.Phase == "playing"; second++ {
		if err := engine.RunBotTurn(game.PlayerOne); err != nil {
			log.Printf("player one bot: %v", err)
		}
		if err := engine.RunBotTurn(game.PlayerTwo); err != nil {
			log.Printf("player two bot: %v", err)
		}
		for tick := 0; tick < game.TickRate; tick++ {
			engine.Tick()
		}
	}
	fmt.Printf("%s\n", engine)
	fmt.Printf("P1 castle HP: %d | units: %d\n", castleHP(engine, game.PlayerOne), len(engine.State.Players[game.PlayerOne].Units))
	fmt.Printf("P2 castle HP: %d | units: %d\n", castleHP(engine, game.PlayerTwo), len(engine.State.Players[game.PlayerTwo].Units))

	if engine.State.Phase == "finished" {
		fmt.Printf("Result: winner=%s reason=%s\n", engine.State.Winner, engine.State.EndReason)
	}
}

func castleHP(engine *game.Engine, playerID string) int {
	for _, building := range engine.State.Players[playerID].Buildings {
		if building.Type == "castle" {
			return building.HP
		}
	}
	return 0
}