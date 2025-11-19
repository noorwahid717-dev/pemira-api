package ws

import (
	"context"
	"log/slog"
	"sync"

	"nhooyr.io/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

type Client struct {
	Conn    *websocket.Conn
	Channel string
	Send    chan Message
}

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.Channel] == nil {
				h.clients[client.Channel] = make(map[*Client]bool)
			}
			h.clients[client.Channel][client] = true
			h.mu.Unlock()
			slog.Info("client registered", "channel", client.Channel)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.Channel]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.Channel)
					}
				}
			}
			h.mu.Unlock()
			slog.Info("client unregistered", "channel", client.Channel)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.Channel]
			h.mu.RUnlock()
			
			for client := range clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					h.mu.Lock()
					delete(h.clients[message.Channel], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}
