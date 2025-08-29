package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"stellarsync-server/internal/models"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	server *Server
	mu     sync.Mutex
	userID string
	name   string
}

// MessageHandlerFunc is a function type for handling messages
type MessageHandlerFunc func(*Client, models.Message)

// Server represents the WebSocket server
type Server struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	handler    MessageHandlerFunc
	userData   map[string]interface{} // Store latest character data per user
}

// NewServer creates a new WebSocket server
func NewServer(handler MessageHandlerFunc) *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		handler:    handler,
		userData:   make(map[string]interface{}),
	}
}

// Start starts the server
func (s *Server) Start() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(s.clients))

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(s.clients))

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		server: s,
	}

	s.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the server
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(100 * 1024 * 1024) // 100MB max message size for file transfers
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse the message
		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Handle the message using the message handler
		if c.server.handler != nil {
			c.server.handler(c, msg)
		}
	}
}

// writePump pumps messages from the server to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetServer returns the server instance
func (c *Client) GetServer() *Server {
	return c.server
}

// GetUserID returns the user ID
func (c *Client) GetUserID() string {
	return c.userID
}

// GetName returns the user name
func (c *Client) GetName() string {
	return c.name
}

// SetUserInfo sets the user information
func (c *Client) SetUserInfo(userID, name string) {
	c.userID = userID
	c.name = name
}

// GetOnlineUsers returns a list of online users
func (s *Server) GetOnlineUsers() []models.UserInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []models.UserInfo
	for client := range s.clients {
		users = append(users, models.UserInfo{
			ID:       client.userID,
			Name:     client.name,
			Online:   true,
			LastSeen: time.Now().Unix(),
		})
	}
	return users
}

// StoreUserData stores character data for a user
func (s *Server) StoreUserData(userID string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userData[userID] = data
}

// GetUserData retrieves character data for a user
func (s *Server) GetUserData(userID string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.userData[userID]
	return data, exists
}

// SendMessage sends a message to this client
func (c *Client) SendMessage(msg models.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.server.clients, c)
	}
}

// BroadcastToOthers broadcasts a message to all other clients
func (s *Server) BroadcastToOthers(sender ClientInterface, msg models.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	s.mu.RLock()
	for client := range s.clients {
		if client != sender {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
	s.mu.RUnlock()
}
