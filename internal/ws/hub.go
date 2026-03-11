// Package ws implements a WebSocket hub for real-time task messaging.
package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

// IncomingMessage represents a message sent from client to server.
type IncomingMessage struct {
	Type    string    `json:"type"`
	TaskID  uuid.UUID `json:"task_id,omitempty"`
	Content string    `json:"content,omitempty"`
}

// OutgoingMessage represents a message sent from server to client.
type OutgoingMessage struct {
	Type      string      `json:"type"`
	TaskID    uuid.UUID   `json:"task_id,omitempty"`
	SessionID uuid.UUID   `json:"session_id,omitempty"`
	Message   interface{} `json:"message,omitempty"`
	Session   interface{} `json:"session,omitempty"`
	Data      string      `json:"data,omitempty"`
	ExitCode  *int        `json:"exit_code,omitempty"`
	Status    string      `json:"status,omitempty"`
}

// MessageHandler is called when a client sends a chat message via WebSocket.
// The handler should persist the message and return the saved model for broadcast.
type MessageHandler func(userID uuid.UUID, taskID uuid.UUID, content string) (interface{}, error)

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID uuid.UUID
	send   chan []byte
	rooms  map[uuid.UUID]bool // task IDs this client is subscribed to
	mu     sync.Mutex
}

// Hub manages WebSocket connections organized by task rooms.
type Hub struct {
	clients    map[*Client]bool
	rooms      map[uuid.UUID]map[*Client]bool // taskID -> set of clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan roomMessage
	mu         sync.RWMutex

	// OnMessage is called when a client sends a chat message.
	// Set this before calling Run().
	OnMessage MessageHandler
}

type roomMessage struct {
	taskID uuid.UUID
	data   []byte
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan roomMessage, 256),
	}
}

// Run starts the hub's main event loop. Call this in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Remove from all rooms
				for taskID := range client.rooms {
					if room, ok := h.rooms[taskID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.rooms, taskID)
						}
					}
				}
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room, ok := h.rooms[msg.taskID]
			if ok {
				for client := range room {
					select {
					case client.send <- msg.data:
					default:
						// Client buffer full — disconnect
						go func(c *Client) { h.unregister <- c }(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToTask sends a message to all clients subscribed to a task.
func (h *Hub) BroadcastToTask(taskID uuid.UUID, msg *OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws: marshal broadcast", "error", err)
		return
	}
	h.broadcast <- roomMessage{taskID: taskID, data: data}
}

// ServeWS upgrades an HTTP connection to WebSocket and registers the client.
func (h *Hub) ServeWS(conn *websocket.Conn, userID uuid.UUID) {
	client := &Client{
		hub:    h,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
		rooms:  make(map[uuid.UUID]bool),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *Hub) subscribe(client *Client, taskID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[taskID]
	if !ok {
		room = make(map[*Client]bool)
		h.rooms[taskID] = room
	}
	room[client] = true
	client.mu.Lock()
	client.rooms[taskID] = true
	client.mu.Unlock()

	slog.Debug("ws: client subscribed", "user_id", client.userID, "task_id", taskID)
}

func (h *Hub) unsubscribe(client *Client, taskID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[taskID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, taskID)
		}
	}
	client.mu.Lock()
	delete(client.rooms, taskID)
	client.mu.Unlock()

	slog.Debug("ws: client unsubscribed", "user_id", client.userID, "task_id", taskID)
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
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("ws: read error", "error", err)
			}
			return
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("ws: invalid message", "error", err)
			continue
		}

		switch msg.Type {
		case "subscribe":
			if msg.TaskID == uuid.Nil {
				continue
			}
			c.hub.subscribe(c, msg.TaskID)

		case "unsubscribe":
			if msg.TaskID == uuid.Nil {
				continue
			}
			c.hub.unsubscribe(c, msg.TaskID)

		case "message":
			if msg.TaskID == uuid.Nil || msg.Content == "" {
				continue
			}
			if c.hub.OnMessage == nil {
				slog.Warn("ws: no message handler configured")
				continue
			}
			saved, err := c.hub.OnMessage(c.userID, msg.TaskID, msg.Content)
			if err != nil {
				slog.Error("ws: handle message", "error", err)
				continue
			}
			c.hub.BroadcastToTask(msg.TaskID, &OutgoingMessage{
				Type:    "message",
				TaskID:  msg.TaskID,
				Message: saved,
			})

		default:
			slog.Warn("ws: unknown message type", "type", msg.Type)
		}
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
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
