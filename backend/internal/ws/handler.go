package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"kingdom-arena/internal/room"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	Hub  *Hub
	Room *room.Room
	mu   sync.Mutex
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	roomMgr    *room.Manager
	mu         sync.RWMutex
}

func NewHub(roomMgr *room.Manager) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		roomMgr:    roomMgr,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("[Hub] Client connected: %s (total: %d)", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("[Hub] Client disconnected: %s (total: %d)", client.ID, len(h.clients))
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Hub] Upgrade error: %v", err)
		return
	}

	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		clientID = "guest_" + r.RemoteAddr
	}

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  h,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		log.Printf("[WS] Received from %s: %s", c.ID, string(message))
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		c.mu.Unlock()
		if err != nil {
			break
		}
	}
}

func (c *Client) SendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.Send <- data:
	default:
	}
}
