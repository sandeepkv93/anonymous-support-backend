package websocket

import (
	"sync"
	"time"
)

type Hub struct {
	clients    map[string]*Client
	broadcast  chan WSMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan WSMessage, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()

			h.BroadcastUserOnline(client.userID, client.username)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

			h.BroadcastUserOffline(client.userID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				client.SendMessage(message)
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID string, msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		client.SendMessage(msg)
	}
}

func (h *Hub) Broadcast(msg WSMessage) {
	h.broadcast <- msg
}

func (h *Hub) BroadcastUserOnline(userID, username string) {
	msg := WSMessage{
		Type:      WSMessageTypeUserOnline,
		Timestamp: time.Now(),
	}
	h.Broadcast(msg)
}

func (h *Hub) BroadcastUserOffline(userID string) {
	msg := WSMessage{
		Type:      WSMessageTypeUserOffline,
		Timestamp: time.Now(),
	}
	h.Broadcast(msg)
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
