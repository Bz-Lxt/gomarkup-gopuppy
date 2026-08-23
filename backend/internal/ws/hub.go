package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gopuppy/internal/auth"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // origin already gated by CORS on HTTP; WS token is the real gate
	},
}

type client struct {
	userID   uuid.UUID
	familyID uuid.UUID
	conn     *websocket.Conn
	send     chan []byte
}

type Hub struct {
	log      *slog.Logger
	issuer   *auth.Issuer
	families *repo.Families
	mu       sync.RWMutex
	rooms    map[uuid.UUID]map[*client]struct{}
}

func New(log *slog.Logger, issuer *auth.Issuer, families *repo.Families) *Hub {
	return &Hub{
		log:      log,
		issuer:   issuer,
		families: families,
		rooms:    map[uuid.UUID]map[*client]struct{}{},
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	c, err := h.issuer.Parse(token)
	if err != nil || c.Kind != "access" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	familyID, err := uuid.Parse(r.URL.Query().Get("family_id"))
	if err != nil {
		http.Error(w, "bad family", http.StatusUnprocessableEntity)
		return
	}
	if _, err := h.families.Member(r.Context(), familyID, c.UserID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("ws upgrade", "err", err)
		return
	}
	cl := &client{userID: c.UserID, familyID: familyID, conn: conn, send: make(chan []byte, 16)}
	h.add(cl)
	go cl.write()
	cl.read(h)
}

func (h *Hub) add(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.familyID] == nil {
		h.rooms[c.familyID] = map[*client]struct{}{}
	}
	h.rooms[c.familyID][c] = struct{}{}
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[c.familyID]; ok {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, c.familyID)
		}
	}
	_ = c.conn.Close()
}

func (h *Hub) Broadcast(familyID uuid.UUID, msg domain.WSMessage) {
	msg.FamilyID = familyID
	if msg.At.IsZero() {
		msg.At = clock.Now()
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[familyID] {
		select {
		case c.send <- b:
		default:
		}
	}
}

func (c *client) read(h *Hub) {
	defer h.remove(c)
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *client) write() {
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-tick.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
