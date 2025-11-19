package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/ws/{channel}", h.HandleWebSocket)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	channel := chi.URLParam(r, "channel")
	
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("failed to accept websocket", "error", err)
		return
	}

	client := &Client{
		Conn:    conn,
		Channel: channel,
		Send:    make(chan Message, 256),
	}

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go h.writePump(ctx, client)
	h.readPump(ctx, client)
}

func (h *Handler) readPump(ctx context.Context, client *Client) {
	defer client.Conn.Close(websocket.StatusNormalClosure, "")

	for {
		_, _, err := client.Conn.Read(ctx)
		if err != nil {
			slog.Error("websocket read error", "error", err)
			return
		}
	}
}

func (h *Handler) writePump(ctx context.Context, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.Close(websocket.StatusNormalClosure, "")
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				slog.Error("failed to marshal message", "error", err)
				continue
			}

			if err := client.Conn.Write(ctx, websocket.MessageText, data); err != nil {
				slog.Error("websocket write error", "error", err)
				return
			}

		case <-ticker.C:
			if err := client.Conn.Ping(ctx); err != nil {
				return
			}
		}
	}
}
