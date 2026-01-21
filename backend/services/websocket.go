package services

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WsEventType represents WebSocket event types
type WsEventType string

const (
	WsEventProgress     WsEventType = "progress"
	WsEventVideoCreated WsEventType = "video_created"
	WsEventVideoDeleted WsEventType = "video_deleted"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type WsEvent struct {
	Type    WsEventType `json:"type"`
	Payload any         `json:"payload"`
}

type WebSocketService struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WsEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WsEvent),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (s *WebSocketService) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			log.Println("WebSocket client registered")

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Close()
				log.Println("WebSocket client unregistered")
			}
			s.mu.Unlock()

		case event := <-s.broadcast:
			s.mu.Lock()
			for client := range s.clients {
				err := client.WriteJSON(event)
				if err != nil {
					log.Printf("WebSocket error: %v", err)
					client.Close()
					delete(s.clients, client)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *WebSocketService) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	s.register <- conn

	// Keep connection alive and listen for close
	go func() {
		defer func() {
			s.unregister <- conn
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func (s *WebSocketService) Broadcast(eventType WsEventType, payload interface{}) {
	s.broadcast <- WsEvent{
		Type:    eventType,
		Payload: payload,
	}
}
