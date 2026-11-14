package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"kingdom-arena/internal/game"
	"kingdom-arena/internal/room"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Development prototype; restrict this to trusted origins in production.
	},
}

type Client struct {
	ID    string
	Side  string
	Conn  *websocket.Conn
	Send  chan []byte
	Done  chan struct{}
	Hub   *Hub
	mu    sync.Mutex
	once  sync.Once
}

type Hub struct {
	clients       map[string]*Client
	register      chan *Client
	unregister    chan *Client
	roomMgr       *room.Manager
	engine        *game.Engine
	gameMu        sync.Mutex
	mu            sync.RWMutex
	reservedSides map[string]bool
}

type clientEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type clientPayload struct {
	UnitType   string `json:"unitType"`
	Col        int    `json:"col"`
	Row        int    `json:"row"`
	UnitID     string `json:"unitId"`
	TargetID   string `json:"targetId"`
	BuildingID string `json:"buildingId"`
	CommandID  string `json:"commandId"`
	Timestamp  int64  `json:"timestamp"`
}

func NewHub(roomMgr *room.Manager) *Hub {
	return &Hub{
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		roomMgr:       roomMgr,
		engine:        game.NewEngineFactory().CreateEngine(),
		reservedSides: make(map[string]bool),
	}
}

func (h *Hub) reserveSide(requested string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if (requested == game.PlayerOne || requested == game.PlayerTwo) && !h.reservedSides[requested] {
		h.reservedSides[requested] = true
		return requested
	}
	for _, side := range []string{game.PlayerOne, game.PlayerTwo} {
		if !h.reservedSides[side] {
			h.reservedSides[side] = true
			return side
		}
	}
	return ""
}

func (h *Hub) Run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			count := len(h.clients)
			h.mu.Unlock()

			h.gameMu.Lock()
			snapshot, err := h.engine.SnapshotFor(client.Side)
			h.gameMu.Unlock()
			if err == nil {
				client.SendMessage(map[string]interface{}{
					"type": "GAME_START",
					"payload": map[string]interface{}{
						"gameState": snapshot,
						"yourSide":  client.Side,
					},
				})
			}
			log.Printf("[Hub] %s connected as %s (clients: %d)", client.ID, client.Side, count)

		case client := <-h.unregister:
			h.mu.Lock()
			if current, ok := h.clients[client.ID]; ok && current == client {
				delete(h.clients, client.ID)
			}
			delete(h.reservedSides, client.Side)
			count := len(h.clients)
			h.mu.Unlock()

			client.once.Do(func() { close(client.Done) })
			log.Printf("[Hub] %s disconnected (clients: %d)", client.ID, count)

		case <-ticker.C:
			h.mu.RLock()
			clients := make([]*Client, 0, len(h.clients))
			for _, client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()
			if len(clients) == 0 {
				continue
			}

			h.gameMu.Lock()
			h.engine.Tick()
			for _, client := range clients {
				snapshot, err := h.engine.SnapshotFor(client.Side)
				if err != nil {
					continue
				}
				client.SendMessage(map[string]interface{}{
					"type":    "STATE_UPDATE",
					"payload": map[string]interface{}{"gameState": snapshot},
				})
			}
			h.gameMu.Unlock()
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Hub] Upgrade error: %v", err)
		return
	}

	side := h.reserveSide(r.URL.Query().Get("side"))
	if side == "" {
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Both player slots are occupied"),
		)
		_ = conn.Close()
		return
	}

	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		clientID = fmt.Sprintf("guest_%d", time.Now().UnixNano())
	}
	client := &Client{
		ID:   clientID,
		Side: side,
		Conn: conn,
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
		Hub:  h,
	}

	h.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var envelope clientEnvelope
		if err := json.Unmarshal(message, &envelope); err != nil {
			c.SendMessage(map[string]interface{}{
				"type": "COMMAND_REJECTED",
				"payload": map[string]string{"message": "Invalid message format"},
			})
			continue
		}

		var payload clientPayload
		if len(envelope.Payload) > 0 {
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				c.SendMessage(map[string]interface{}{
					"type": "COMMAND_REJECTED",
					"payload": map[string]string{"message": "Invalid command payload"},
				})
				continue
			}
		}

		if envelope.Type == "PING" {
			c.SendMessage(map[string]interface{}{
				"type": "PONG",
				"payload": map[string]interface{}{
					"timestamp":  payload.Timestamp,
					"serverTime": time.Now().UnixMilli(),
				},
			})
			continue
		}

		command := game.Command{CommandID: payload.CommandID}
		switch envelope.Type {
		case "DEPLOY_UNIT":
			command.Type = "DEPLOY_UNIT"
			command.UnitType, command.Col, command.Row = payload.UnitType, payload.Col, payload.Row
		case "MOVE_TO":
			command.Type = "MOVE_TO"
			command.UnitID, command.Col, command.Row = payload.UnitID, payload.Col, payload.Row
		case "SET_TARGET":
			command.Type = "SET_TARGET"
			command.UnitID, command.TargetID = payload.UnitID, payload.TargetID
		case "UPGRADE":
			command.Type, command.BuildingID = "UPGRADE", payload.BuildingID
		case "SURRENDER":
			command.Type = "SURRENDER"
		default:
			c.SendMessage(map[string]interface{}{
				"type": "COMMAND_REJECTED",
				"payload": map[string]string{"message": "Unsupported command"},
			})
			continue
		}

		c.Hub.gameMu.Lock()
		err = c.Hub.engine.ApplyCommand(c.Side, command)
		c.Hub.gameMu.Unlock()
		if err != nil {
			c.SendMessage(map[string]interface{}{
				"type": "COMMAND_REJECTED",
				"payload": map[string]string{"message": err.Error()},
			})
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for {
		select {
		case <-c.Done:
			return
		case message := <-c.Send:
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) SendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case <-c.Done:
		return
	default:
	}
	select {
	case <-c.Done:
	case c.Send <- data:
	default:
	}
}
