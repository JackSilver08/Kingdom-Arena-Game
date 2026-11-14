package game

import (
	"kingdom-arena/pkg/models"
)

type Engine struct {
	State *models.GameState
}

type EngineFactory struct{}

func NewEngineFactory() *EngineFactory {
	return &EngineFactory{}
}

func (f *EngineFactory) CreateEngine() *Engine {
	return &Engine{
		State: models.NewGameState(),
	}
}
