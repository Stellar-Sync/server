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
	zone   string // Current zone/area the user is in
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

	// Sync request queuing system
	syncRequests map[string][]SyncRequest // Map of targetUserID -> pending sync requests
	syncMu       sync.RWMutex
}

// SyncRequest represents a pending sync request
type SyncRequest struct {
	RequestingUserID   string
	RequestingUserName string
	RequestTime        int64
	Status             string // "pending", "fulfilled", "expired"
}

// NewServer creates a new WebSocket server
func NewServer(handler MessageHandlerFunc) *Server {
	return &Server{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		handler:      handler,
		userData:     make(map[string]interface{}),
		syncRequests: make(map[string][]SyncRequest),
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
				// Clean up user data when they disconnect
				userID := client.GetUserID()
				if userID != "" {
					if _, exists := s.userData[userID]; exists {
						log.Printf("[DISCONNECT] User %s disconnected, but keeping their data for future requests", userID)
						// Note: We're keeping the data for now, but you can uncomment the next line to clear it
						delete(s.userData, userID)
					}
				}

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

// GetZone returns the user's current zone
func (c *Client) GetZone() string {
	return c.zone
}

// SetUserInfo sets the user information
func (c *Client) SetUserInfo(userID, name, zone string) {
	c.userID = userID
	c.name = name
	c.zone = zone
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
			Zone:     client.zone,
			Online:   true,
			LastSeen: time.Now().Unix(),
		})
	}
	return users
}

// StoreUserData stores character data for a user
func (s *Server) StoreUserData(userID string, data interface{}) {
	// Add timestamp to track when data was stored
	dataWithMeta := map[string]interface{}{
		"data":      data,
		"stored_at": time.Now().Unix(),
		"user_id":   userID,
	}

	// Store the data (with lock)
	s.mu.Lock()
	s.userData[userID] = dataWithMeta
	s.mu.Unlock()
	log.Printf("[DATA_STORE] Stored data for user %s at %d", userID, time.Now().Unix())

	// Check if there are any pending sync requests for this user's data (without holding the lock)
	s.checkAndFulfillPendingSyncRequests(userID, dataWithMeta)
}

// ListAllUserData returns information about all stored user data (for debugging)
func (s *Server) ListAllUserData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for userID, dataWithMeta := range s.userData {
		if dataMap, ok := dataWithMeta.(map[string]interface{}); ok {
			result[userID] = map[string]interface{}{
				"stored_at": dataMap["stored_at"],
				"has_data":  dataMap["data"] != nil,
			}
		} else {
			result[userID] = "legacy_format"
		}
	}
	return result
}

// GetUserData retrieves character data for a user
func (s *Server) GetUserData(userID string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dataWithMeta, exists := s.userData[userID]
	if !exists {
		return nil, false
	}

	// Extract the actual data from the metadata wrapper
	if dataMap, ok := dataWithMeta.(map[string]interface{}); ok {
		if actualData, hasData := dataMap["data"]; hasData {
			log.Printf("[DATA_GET] Retrieved data for user %s (stored at %v)", userID, dataMap["stored_at"])
			return actualData, true
		}
	}

	// Fallback for old data format
	log.Printf("[DATA_GET] Retrieved legacy data format for user %s", userID)
	return dataWithMeta, true
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

	log.Printf("[SEND] Sending message to client %s (%s): type=%s, size=%d bytes", c.GetUserID(), c.GetName(), msg.Type, len(data))

	select {
	case c.send <- data:
		log.Printf("[SEND] Message queued successfully for client %s", c.GetUserID())
	default:
		log.Printf("[SEND] Failed to queue message for client %s - channel full or closed", c.GetUserID())
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

// checkAndFulfillPendingSyncRequests checks if there are pending sync requests for a user's data
func (s *Server) checkAndFulfillPendingSyncRequests(userID string, dataWithMeta interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[SYNC_QUEUE] PANIC in checkAndFulfillPendingSyncRequests: %v", r)
		}
	}()

	s.syncMu.Lock()
	pendingRequests, exists := s.syncRequests[userID]
	if !exists || len(pendingRequests) == 0 {
		s.syncMu.Unlock()
		return // No pending requests
	}

	log.Printf("[SYNC_QUEUE] Found %d pending sync requests for user %s, fulfilling them", len(pendingRequests), userID)

	// Debug: Log the details of each request
	for i, request := range pendingRequests {
		log.Printf("[SYNC_QUEUE] Request %d: %s (%s) wants data from %s, status=%s, time=%d",
			i+1, request.RequestingUserName, request.RequestingUserID, userID, request.Status, request.RequestTime)
	}

	// Clear the fulfilled requests before processing to avoid deadlock
	delete(s.syncRequests, userID)
	s.syncMu.Unlock()
	log.Printf("[SYNC_QUEUE] Cleared %d fulfilled sync requests for user %s", len(pendingRequests), userID)

	// Fulfill all pending requests (now without holding syncMu to avoid deadlock)
	for _, request := range pendingRequests {
		if request.Status == "pending" {
			log.Printf("[SYNC_QUEUE] Processing pending request from %s", request.RequestingUserName)
			s.fulfillSyncRequest(request, userID, dataWithMeta)
		} else {
			log.Printf("[SYNC_QUEUE] Skipping non-pending request from %s (status: %s)", request.RequestingUserName, request.Status)
		}
	}
}

// fulfillSyncRequest sends character data to a user who was waiting for it
func (s *Server) fulfillSyncRequest(request SyncRequest, targetUserID string, dataWithMeta interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[SYNC_QUEUE] PANIC in fulfillSyncRequest: %v", r)
		}
	}()

	log.Printf("[SYNC_QUEUE] Starting fulfillment for request: %s wants data from %s", request.RequestingUserName, targetUserID)

	// Find the requesting client
	log.Printf("[SYNC_QUEUE] Step 1: About to acquire read lock")
	var requestingClient *Client
	s.mu.RLock()
	log.Printf("[SYNC_QUEUE] Step 2: Acquired read lock, iterating through clients")
	clientCount := 0
	for client := range s.clients {
		clientCount++
		log.Printf("[SYNC_QUEUE] Step 3: Checking client %d: %s", clientCount, client.GetUserID())
		if client.GetUserID() == request.RequestingUserID {
			requestingClient = client
			log.Printf("[SYNC_QUEUE] Found requesting client %s (%s) in %d total clients", request.RequestingUserID, request.RequestingUserName, clientCount)
			break
		}
	}
	log.Printf("[SYNC_QUEUE] Step 4: Finished iterating through %d clients, releasing lock", clientCount)
	s.mu.RUnlock()
	log.Printf("[SYNC_QUEUE] Step 5: Released read lock")

	if requestingClient == nil {
		log.Printf("[SYNC_QUEUE] Requesting client %s not found in %d total clients, skipping fulfillment", request.RequestingUserID, clientCount)
		return
	}

	log.Printf("[SYNC_QUEUE] Step 6: Client found, extracting data from metadata")
	// Extract the actual data from metadata
	var actualData interface{}
	if dataMap, ok := dataWithMeta.(map[string]interface{}); ok {
		actualData = dataMap["data"]
		log.Printf("[SYNC_QUEUE] Step 7: Extracted data from metadata wrapper")
	} else {
		actualData = dataWithMeta // Fallback for legacy format
		log.Printf("[SYNC_QUEUE] Step 7: Using fallback data format")
	}

	log.Printf("[SYNC_QUEUE] Step 8: Getting source character name")
	// Get the source character name
	var sourceCharacterName string
	if charData, ok := actualData.(map[string]interface{}); ok {
		if name, ok := charData["character_name"].(string); ok {
			sourceCharacterName = name
			log.Printf("[SYNC_QUEUE] Step 9: Found source character name: %s", sourceCharacterName)
		} else if name, ok := charData["name"].(string); ok {
			sourceCharacterName = name
			log.Printf("[SYNC_QUEUE] Step 9: Found source character name (fallback): %s", sourceCharacterName)
		} else {
			log.Printf("[SYNC_QUEUE] Step 9: Could not find character name in data")
		}
	} else {
		log.Printf("[SYNC_QUEUE] Step 9: Could not extract character data, type: %T", actualData)
	}

	log.Printf("[SYNC_QUEUE] Step 10: Creating response message")
	// Send the character data to the requesting user
	response := models.Message{
		Type: "user_character_data",
		Data: map[string]interface{}{
			"user_id":               targetUserID,
			"source_character_name": sourceCharacterName,
			"data":                  actualData,
		},
	}

	log.Printf("[SYNC_QUEUE] Step 11: Sending response to client %s: type=%s, source=%s", request.RequestingUserID, response.Type, sourceCharacterName)
	requestingClient.SendMessage(response)
	log.Printf("[SYNC_QUEUE] Step 12: Fulfilled pending sync request: sent data from %s to %s", sourceCharacterName, request.RequestingUserName)
}

// QueueSyncRequest adds a sync request to the queue when data isn't available yet
func (s *Server) QueueSyncRequest(targetUserID, requestingUserID, requestingUserName string) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	request := SyncRequest{
		RequestingUserID:   requestingUserID,
		RequestingUserName: requestingUserName,
		RequestTime:        time.Now().Unix(),
		Status:             "pending",
	}

	if s.syncRequests[targetUserID] == nil {
		s.syncRequests[targetUserID] = make([]SyncRequest, 0)
	}

	s.syncRequests[targetUserID] = append(s.syncRequests[targetUserID], request)
	log.Printf("[SYNC_QUEUE] Queued sync request: %s wants data from %s", requestingUserName, targetUserID)
}

// GetPendingSyncRequests returns pending sync requests for a user (for debugging)
func (s *Server) GetPendingSyncRequests(targetUserID string) []SyncRequest {
	s.syncMu.RLock()
	defer s.syncMu.RUnlock()

	if requests, exists := s.syncRequests[targetUserID]; exists {
		return requests
	}
	return []SyncRequest{}
}
