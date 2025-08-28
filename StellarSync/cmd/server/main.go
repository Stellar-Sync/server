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
		handleConnect(client)
	case "character_data":
		handleCharacterData(client, msg)
	default:
		handleUnknownMessage(client, msg)
	}
}

// handleConnect handles client connection messages
func handleConnect(client *websocket.Client) {
	response := models.Message{
		Type: "connected",
		Data: map[string]interface{}{
			"message": "Welcome to Stellar Sync Server!",
			"time":    time.Now().Unix(),
		},
	}
	client.SendMessage(response)
}

// handleCharacterData handles character data messages
func handleCharacterData(client *websocket.Client, msg models.Message) {
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
