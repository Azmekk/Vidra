package services

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WsEventType string

const (
	WsEventVideoCreated WsEventType = "video_created"
	WsEventVideoUpdated WsEventType = "video_updated"
	WsEventVideoDeleted WsEventType = "video_deleted"
	WsEventFileProgress WsEventType = "file_progress"
	WsEventBackupStatus WsEventType = "backup_status"
)

const (
	wsSendBuffer = 64
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = 50 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

type WsEvent struct {
	Type    WsEventType `json:"type"`
	Payload any         `json:"payload"`
}

type wsClient struct {
	conn *websocket.Conn
	send chan WsEvent
}

// WebSocketService fans events out to connected clients. Each client has its
// own buffered queue so a slow phone never stalls a download.
type WebSocketService struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{clients: map[*wsClient]struct{}{}}
}

func (s *WebSocketService) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	c := &wsClient{conn: conn, send: make(chan WsEvent, wsSendBuffer)}
	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()

	go s.writeLoop(c)
	go s.readLoop(c)
}

func (s *WebSocketService) Broadcast(eventType WsEventType, payload any) {
	event := WsEvent{Type: eventType, Payload: payload}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for c := range s.clients {
		select {
		case c.send <- event:
		default:
			go s.remove(c)
		}
	}
}

func (s *WebSocketService) remove(c *wsClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clients[c]; ok {
		delete(s.clients, c)
		close(c.send)
	}
}

func (s *WebSocketService) readLoop(c *wsClient) {
	defer s.remove(c)
	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *WebSocketService) writeLoop(c *wsClient) {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case event, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteJSON(event); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
