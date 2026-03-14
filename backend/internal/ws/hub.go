package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512 * 1024 // 512 KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// Origin validation is handled by CORS middleware upstream.
		return true
	},
}

// Envelope is the universal WS message wrapper (both directions).
type Envelope struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Channel string          `json:"channel,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// BroadcastMsg is an internal message to push to a channel.
type BroadcastMsg struct {
	Channel  string
	Envelope Envelope
}

// Hub manages all connected WebSocket clients and channel subscriptions.
type Hub struct {
	mu        sync.RWMutex
	clients   map[*Client]bool
	channels  map[string]map[*Client]bool // channel → subscriber set

	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMsg
}

// New creates and returns a Hub. Call Run() in a goroutine.
func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		channels:   make(map[string]map[*Client]bool),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan BroadcastMsg, 512),
	}
}

// Run starts the hub event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if h.clients[client] {
				delete(h.clients, client)
				// Remove from all channels.
				for ch, subs := range h.channels {
					delete(subs, client)
					if len(subs) == 0 {
						delete(h.channels, ch)
					}
				}
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			subs := h.channels[msg.Channel]
			targets := make([]*Client, 0, len(subs))
			for c := range subs {
				targets = append(targets, c)
			}
			h.mu.RUnlock()

			data, err := json.Marshal(msg.Envelope)
			if err != nil {
				log.Printf("ws: marshal broadcast: %v", err)
				continue
			}
			for _, c := range targets {
				select {
				case c.send <- data:
				default:
					// Client send buffer full — drop and unregister.
					h.unregister <- c
				}
			}
		}
	}
}

// Broadcast sends an envelope to all subscribers of the given channel.
func (h *Hub) Broadcast(channel string, env Envelope) {
	h.broadcast <- BroadcastMsg{Channel: channel, Envelope: env}
}

// Subscribe adds a client to a channel.
func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true
}

// Unsubscribe removes a client from a channel.
func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.channels[channel]; ok {
		delete(subs, client)
		if len(subs) == 0 {
			delete(h.channels, channel)
		}
	}
}

// Client is a single WebSocket connection.
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	UserID        string
	subscriptions map[string]bool
	mu            sync.Mutex
}

// ServeWS upgrades an HTTP connection to WebSocket and registers the client.
// The caller must have already validated auth and set userID.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade: %v", err)
		return
	}

	client := &Client{
		hub:           h,
		conn:          conn,
		send:          make(chan []byte, 256),
		UserID:        userID,
		subscriptions: make(map[string]bool),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: read: %v", err)
			}
			return
		}
		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		var env Envelope
		if err := json.Unmarshal(msg, &env); err != nil {
			c.sendError(env.ID, "PROTOCOL_ERROR", "malformed message")
			continue
		}

		c.handleMessage(env)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(env Envelope) {
	switch env.Type {
	case "subscribe":
		var p struct {
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.Channel == "" {
			c.sendError(env.ID, "VALIDATION_ERROR", "channel is required")
			return
		}
		c.hub.Subscribe(c, p.Channel)
		c.mu.Lock()
		c.subscriptions[p.Channel] = true
		c.mu.Unlock()
		c.sendEnvelope(Envelope{
			Type:    "subscribed",
			ID:      env.ID,
			Payload: mustMarshal(map[string]string{"channel": p.Channel}),
		})

	case "unsubscribe":
		var p struct {
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.Channel == "" {
			c.sendError(env.ID, "VALIDATION_ERROR", "channel is required")
			return
		}
		c.hub.Unsubscribe(c, p.Channel)
		c.mu.Lock()
		delete(c.subscriptions, p.Channel)
		c.mu.Unlock()
		c.sendEnvelope(Envelope{
			Type:    "unsubscribed",
			ID:      env.ID,
			Payload: mustMarshal(map[string]string{"channel": p.Channel}),
		})

	case "ping":
		c.sendEnvelope(Envelope{
			Type:    "pong",
			Payload: mustMarshal(map[string]string{"ts": time.Now().UTC().Format(time.RFC3339Nano)}),
		})

	case "presence.update":
		// Handled by presence tracker — stub for now, wired in Phase 4.
		// No-op here; presence tracker will hook into this via hub extension.

	case "agent.input":
		// Handled by agent process manager — stub for now, wired in Phase 3.
		c.sendError(env.ID, "NOT_IMPLEMENTED", "agent input via WebSocket not yet available")

	default:
		c.sendError(env.ID, "PROTOCOL_ERROR", "unknown message type: "+env.Type)
	}
}

func (c *Client) sendEnvelope(env Envelope) {
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendError(id, code, message string) {
	c.sendEnvelope(Envelope{
		Type: "error",
		ID:   id,
		Payload: mustMarshal(map[string]string{
			"code":    code,
			"message": message,
		}),
	})
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
