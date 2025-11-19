package tps

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	
	"pemira-api/internal/shared/ctxkeys"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type WSCheckinEvent struct {
	CheckinID int64      `json:"checkin_id"`
	Voter     *VoterInfo `json:"voter,omitempty"`
	Status    string     `json:"status,omitempty"`
	ScanAt    *time.Time `json:"scan_at,omitempty"`
}

type WSHub struct {
	clients    map[int64]map[*websocket.Conn]bool // tpsID -> connections
	broadcast  chan BroadcastMessage
	register   chan Registration
	unregister chan Registration
	mu         sync.RWMutex
}

type Registration struct {
	tpsID int64
	conn  *websocket.Conn
}

type BroadcastMessage struct {
	tpsID   int64
	message WSMessage
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[int64]map[*websocket.Conn]bool),
		broadcast:  make(chan BroadcastMessage, 256),
		register:   make(chan Registration),
		unregister: make(chan Registration),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.mu.Lock()
			if h.clients[reg.tpsID] == nil {
				h.clients[reg.tpsID] = make(map[*websocket.Conn]bool)
			}
			h.clients[reg.tpsID][reg.conn] = true
			h.mu.Unlock()
			
		case reg := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[reg.tpsID]; ok {
				if _, ok := clients[reg.conn]; ok {
					delete(clients, reg.conn)
					reg.conn.Close()
					if len(clients) == 0 {
						delete(h.clients, reg.tpsID)
					}
				}
			}
			h.mu.Unlock()
			
		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.tpsID]
			h.mu.RUnlock()
			
			data, _ := json.Marshal(msg.message)
			for conn := range clients {
				err := conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					h.unregister <- Registration{tpsID: msg.tpsID, conn: conn}
				}
			}
		}
	}
}

func (h *WSHub) BroadcastCheckinNew(tpsID, checkinID int64, voter *VoterInfo, scanAt time.Time) {
	h.broadcast <- BroadcastMessage{
		tpsID: tpsID,
		message: WSMessage{
			Type: "CHECKIN_NEW",
			Data: WSCheckinEvent{
				CheckinID: checkinID,
				Voter:     voter,
				ScanAt:    &scanAt,
			},
		},
	}
}

func (h *WSHub) BroadcastCheckinUpdated(tpsID, checkinID int64, status string) {
	h.broadcast <- BroadcastMessage{
		tpsID: tpsID,
		message: WSMessage{
			Type: "CHECKIN_UPDATED",
			Data: WSCheckinEvent{
				CheckinID: checkinID,
				Status:    status,
			},
		},
	}
}

type WSHandler struct {
	hub     *WSHub
	service *Service
}

func NewWSHandler(hub *WSHub, service *Service) *WSHandler {
	return &WSHandler{
		hub:     hub,
		service: service,
	}
}

func (h *WSHandler) RegisterRoutes(r chi.Router) {
	r.Get("/ws/tps/{tps_id}/queue", h.HandleTPSQueue)
}

func (h *WSHandler) HandleTPSQueue(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tps_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid TPS ID", http.StatusBadRequest)
		return
	}
	
	// Verify auth
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Verify TPS access
	hasAccess, err := h.service.repo.IsPanitiaAssigned(r.Context(), tpsID, userID)
	if err != nil || !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	// Register client
	h.hub.register <- Registration{tpsID: tpsID, conn: conn}
	
	// Send initial queue state
	go h.sendInitialQueue(conn, tpsID)
	
	// Keep connection alive and handle client messages
	go h.readPump(conn, tpsID)
}

func (h *WSHandler) sendInitialQueue(conn *websocket.Conn, tpsID int64) {
	// Note: This function sends initial queue state when client connects
	// Implementation can be enhanced to fetch and send current pending checkins
	
	// Send welcome message
	msg := WSMessage{
		Type: "CONNECTED",
		Data: map[string]interface{}{
			"tps_id":    tpsID,
			"timestamp": time.Now(),
		},
	}
	
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func (h *WSHandler) readPump(conn *websocket.Conn, tpsID int64) {
	defer func() {
		h.hub.unregister <- Registration{tpsID: tpsID, conn: conn}
	}()
	
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// GetHub returns the hub instance for triggering broadcasts from service layer
func (h *WSHandler) GetHub() *WSHub {
	return h.hub
}
