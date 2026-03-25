package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)

type Hub struct {
	// Map of userID -> Connection
	clients map[string]*websocket.Conn
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = conn
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.clients[userID]; ok {
		conn.Close()
		delete(h.clients, userID)
	}
}

// SendToUser pushes a JSON payload to a specific user if they are online
func (h *Hub) SendToUser(userID string, event interface{}) {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		payload, _ := json.Marshal(event)
		_ = conn.WriteMessage(websocket.TextMessage, payload)
	}
}
