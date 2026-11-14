package room

import (
	"sync"

	"kingdom-arena/pkg/models"
)

type Room struct {
	ID        string
	Player1   string
	Player2   string
	GameState *models.GameState
	mu        sync.RWMutex
}

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) CreateRoom(id string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	r := &Room{
		ID:        id,
		GameState: models.NewGameState(),
	}
	m.rooms[id] = r
	return r
}

func (m *Manager) GetRoom(id string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[id]
}

func (m *Manager) DeleteRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (m *Manager) ListRooms() []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	return rooms
}
