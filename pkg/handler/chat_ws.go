package handler

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/gorilla/websocket"
)

const (
	chatWriteWait      = 10 * time.Second
	chatPongWait       = 60 * time.Second
	chatPingPeriod     = (chatPongWait * 9) / 10
	chatMaxMessageSize = 512
)

type chatHub struct {
	mu          sync.RWMutex
	byCompanyID map[int64]map[*chatClient]struct{}
}

type chatClient struct {
	hub       *chatHub
	conn      *websocket.Conn
	send      chan []byte
	companyID int64
}

func newChatHub() *chatHub {
	return &chatHub{
		byCompanyID: make(map[int64]map[*chatClient]struct{}),
	}
}

func (h *chatHub) Register(client *chatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.byCompanyID[client.companyID]; !ok {
		h.byCompanyID[client.companyID] = make(map[*chatClient]struct{})
	}
	h.byCompanyID[client.companyID][client] = struct{}{}
}

func (h *chatHub) Unregister(client *chatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.byCompanyID[client.companyID]
	if !ok {
		return
	}
	if _, exists := clients[client]; exists {
		delete(clients, client)
		close(client.send)
	}
	if len(clients) == 0 {
		delete(h.byCompanyID, client.companyID)
	}
}

func (h *chatHub) BroadcastToCompany(companyID int64, event model.ChatRealtimeEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("chat hub marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.byCompanyID[companyID]
	targets := make([]*chatClient, 0, len(clients))
	for client := range clients {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		select {
		case client.send <- payload:
		default:
			h.Unregister(client)
			_ = client.conn.Close()
		}
	}
}

func (c *chatClient) readPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(chatMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(chatPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(chatPongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *chatClient) writePump() {
	ticker := time.NewTicker(chatPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(chatWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(chatWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
