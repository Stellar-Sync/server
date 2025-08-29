package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"stellarsync-server/internal/models"
	"stellarsync-server/internal/websocket"
)

// handleMessage processes incoming messages
func handleMessage(client *websocket.Client, msg models.Message) {
	log.Printf("Received message: %+v", msg)

	switch msg.Type {
	case "connect":
		handleConnect(client, msg)
	case "character_data":
		handleCharacterData(client, msg)
	case "request_users":
		handleRequestUsers(client)
	case "request_user_data":
		handleRequestUserData(client, msg)
	default:
		handleUnknownMessage(client, msg)
	}
}

// handleConnect handles client connection messages
func handleConnect(client *websocket.Client, msg models.Message) {
	// Extract user info from the connect message
	if data, ok := msg.Data.(map[string]interface{}); ok {
		if userID, ok := data["user_id"].(string); ok {
			if name, ok := data["name"].(string); ok {
				client.SetUserInfo(userID, name)
			}
		}
	}

	response := models.Message{
		Type: "connected",
		Data: map[string]interface{}{
			"message": "Welcome to Stellar Sync Server!",
			"time":    time.Now().Unix(),
		},
	}
	client.SendMessage(response)

	// Broadcast updated user list to all clients
	broadcastUserList(client.GetServer())
}

// handleRequestUsers handles requests for online users list
func handleRequestUsers(client *websocket.Client) {
	users := client.GetServer().GetOnlineUsers()
	response := models.Message{
		Type: "users_list",
		Data: users,
	}
	client.SendMessage(response)
}

// handleRequestUserData handles requests for specific user's character data
func handleRequestUserData(client *websocket.Client, msg models.Message) {
	if data, ok := msg.Data.(map[string]interface{}); ok {
		if targetUserID, ok := data["user_id"].(string); ok {
			if characterData, exists := client.GetServer().GetUserData(targetUserID); exists {
				response := models.Message{
					Type: "user_character_data",
					Data: map[string]interface{}{
						"user_id": targetUserID,
						"data":    characterData,
					},
				}
				client.SendMessage(response)
			} else {
				response := models.Message{
					Type:  "error",
					Error: "User data not found",
				}
				client.SendMessage(response)
			}
		}
	}
}

// broadcastUserList broadcasts the current user list to all clients
func broadcastUserList(server *websocket.Server) {
	users := server.GetOnlineUsers()
	broadcastMsg := models.Message{
		Type: "users_list",
		Data: users,
	}
	server.BroadcastToOthers(nil, broadcastMsg)
}

// handleCharacterData handles character data messages
func handleCharacterData(client *websocket.Client, msg models.Message) {
	// Store the character data for this user
	client.GetServer().StoreUserData(client.GetUserID(), msg.Data)

	// Send acknowledgment to the sender
	response := models.Message{
		Type: "character_data_received",
		Data: map[string]interface{}{
			"received_at": time.Now().Unix(),
			"status":      "ok",
		},
	}
	client.SendMessage(response)

	// Broadcast to other clients
	broadcastMsg := models.Message{
		Type: "character_data_broadcast",
		Data: msg.Data,
	}
	client.GetServer().BroadcastToOthers(client, broadcastMsg)
}

// handleUnknownMessage handles unknown message types
func handleUnknownMessage(client *websocket.Client, msg models.Message) {
	response := models.Message{
		Type:  "error",
		Error: fmt.Sprintf("Unknown message type: %s", msg.Type),
	}
	client.SendMessage(response)
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "stellar-sync-server",
	})
}

func main() {
	// Create WebSocket server with message handler
	wsServer := websocket.NewServer(handleMessage)
	go wsServer.Start()

	// Set up HTTP routes
	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	http.HandleFunc("/health", healthCheckHandler)

	// Add a simple status page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Stellar Sync Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 10px; background: #e8f5e8; border: 1px solid #4caf50; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>Stellar Sync Server</h1>
    <div class="status">
        <strong>Status:</strong> Running<br>
        <strong>WebSocket Endpoint:</strong> <code>/ws</code><br>
        <strong>Health Check:</strong> <a href="/health">/health</a>
    </div>
    <p>This server handles WebSocket connections for Stellar Sync clients.</p>
</body>
</html>
`)
	})

	port := ":6000"
	log.Printf("Starting Stellar Sync Server on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", port)
	log.Printf("Health check: http://localhost%s/health", port)
	log.Printf("Status page: http://localhost%s/", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
