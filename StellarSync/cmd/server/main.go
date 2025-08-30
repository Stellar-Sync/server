package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"stellarsync-server/internal/models"
	"stellarsync-server/internal/proxy"
	"stellarsync-server/internal/websocket"
)

// Global file proxy for forwarding requests to file server
var fileProxy *proxy.FileProxy

// handleFileUpload forwards file upload requests to the file server
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	fileProxy.ForwardRequest(w, r)
}

// handleFileDownload forwards file download requests to the file server
func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	fileProxy.ForwardRequest(w, r)
}

// handleFileList forwards file list requests to the file server
func handleFileList(w http.ResponseWriter, r *http.Request) {
	fileProxy.ForwardRequest(w, r)
}

// handleFileMetadataUpload forwards file metadata upload requests to the file server
func handleFileMetadataUpload(w http.ResponseWriter, r *http.Request) {
	fileProxy.ForwardRequest(w, r)
}

// handleFileMetadataDownload forwards file metadata download requests to the file server
func handleFileMetadataDownload(w http.ResponseWriter, r *http.Request) {
	fileProxy.ForwardRequest(w, r)
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	var userName string
	var userID string

	if data, ok := msg.Data.(map[string]interface{}); ok {
		if id, ok := data["user_id"].(string); ok {
			userID = id
		}
		if name, ok := data["name"].(string); ok {
			userName = name
			client.SetUserInfo(userID, name)
		}
	}

	log.Printf("[CONNECT] User %s (%s) connecting to server", userName, userID)

	response := models.Message{
		Type: "connected",
		Data: map[string]interface{}{
			"message": "Welcome to Stellar Sync Server!",
			"time":    time.Now().Unix(),
		},
	}
	client.SendMessage(response)
	log.Printf("[CONNECT] Sent welcome message to user %s", userName)

	// Log current online users
	server := client.GetServer()
	onlineUsers := server.GetOnlineUsers()
	log.Printf("[CONNECT] Total online users after connection: %d", len(onlineUsers))
	for _, user := range onlineUsers {
		log.Printf("[CONNECT] - Online: %s (%s)", user.Name, user.ID)
	}

	// Broadcast updated user list to all clients
	broadcastUserList(server)
	log.Printf("[CONNECT] Broadcasted updated user list after %s connected", userName)
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
	requestingUserID := client.GetUserID()
	requestingUserName := client.GetName()

	if data, ok := msg.Data.(map[string]interface{}); ok {
		if targetUserID, ok := data["user_id"].(string); ok {
			log.Printf("[REQUEST_DATA] User %s (%s) requesting character data from user %s", requestingUserName, requestingUserID, targetUserID)

			if characterData, exists := client.GetServer().GetUserData(targetUserID); exists {
				// Get the source character name from the character data
				var sourceCharacterName string
				if charData, ok := characterData.(map[string]interface{}); ok {
					if name, ok := charData["character_name"].(string); ok {
						sourceCharacterName = name
					} else if name, ok := charData["name"].(string); ok {
						// Fallback to old format
						sourceCharacterName = name
					}
				}

				log.Printf("[REQUEST_DATA] Found character data for %s (requested by %s)", sourceCharacterName, requestingUserName)

				// Log what we're sending back
				if charData, ok := characterData.(map[string]interface{}); ok {
					log.Printf("[REQUEST_DATA] Sending character data to %s:", requestingUserName)

					// Log Glamourer data
					if glamourerData, ok := charData["glamourer_data"].(string); ok {
						glamourerLength := len(glamourerData)
						log.Printf("[REQUEST_DATA] - Glamourer Data: %d characters", glamourerLength)
					}

					// Log Penumbra meta data
					if penumbraMeta, ok := charData["penumbra_meta"].(string); ok {
						metaLength := len(penumbraMeta)
						log.Printf("[REQUEST_DATA] - Penumbra Meta: %d characters", metaLength)
					}

					// Log Penumbra files data
					if penumbraFiles, ok := charData["penumbra_files"].(map[string]interface{}); ok {
						fileCount := len(penumbraFiles)
						totalBase64Size := 0
						for _, base64Data := range penumbraFiles {
							if base64String, ok := base64Data.(string); ok {
								totalBase64Size += len(base64String)
							}
						}
						log.Printf("[REQUEST_DATA] - Penumbra Files: %d files (%.2f MB base64)", fileCount, float64(totalBase64Size)/(1024*1024))
					}

					// Log Penumbra file data (compressed files)
					if penumbraFileData, ok := charData["penumbra_file_data"].(map[string]interface{}); ok {
						compressedFileCount := len(penumbraFileData)
						totalSize := 0
						for _, compressedData := range penumbraFileData {
							if fileBytes, ok := compressedData.([]byte); ok {
								totalSize += len(fileBytes)
							}
						}
						log.Printf("[REQUEST_DATA] - Penumbra File Data: %d compressed files (%.2f MB total)", compressedFileCount, float64(totalSize)/(1024*1024))
					}
				}

				response := models.Message{
					Type: "user_character_data",
					Data: map[string]interface{}{
						"user_id":               targetUserID,
						"source_character_name": sourceCharacterName,
						"data":                  characterData,
					},
				}
				client.SendMessage(response)
				log.Printf("[REQUEST_DATA] Successfully sent character data from %s to %s", sourceCharacterName, requestingUserName)
			} else {
				log.Printf("[REQUEST_DATA] User %s requested data from user %s, but no data found", requestingUserName, targetUserID)
				response := models.Message{
					Type:  "error",
					Error: "User data not found",
				}
				client.SendMessage(response)
			}
		} else {
			log.Printf("[REQUEST_DATA] User %s sent request without user_id", requestingUserName)
		}
	} else {
		log.Printf("[REQUEST_DATA] User %s sent invalid request data format", requestingUserName)
	}
}

// broadcastUserList broadcasts the current user list to all clients
func broadcastUserList(server *websocket.Server) {
	users := server.GetOnlineUsers()
	log.Printf("[BROADCAST] Broadcasting user list to all clients: %d users", len(users))
	for _, user := range users {
		log.Printf("[BROADCAST] - User: %s (%s)", user.Name, user.ID)
	}

	broadcastMsg := models.Message{
		Type: "users_list",
		Data: users,
	}
	server.BroadcastToOthers(nil, broadcastMsg)
	log.Printf("[BROADCAST] User list broadcast completed")
}

// handleCharacterData handles character data messages
func handleCharacterData(client *websocket.Client, msg models.Message) {
	userID := client.GetUserID()
	userName := client.GetName()

	log.Printf("[CHARACTER_DATA] User %s (%s) sending character data", userName, userID)

	// Log the structure of received data
	if data, ok := msg.Data.(map[string]interface{}); ok {
		log.Printf("[CHARACTER_DATA] Data structure for user %s:", userName)

		// Log basic character info
		if name, ok := data["character_name"].(string); ok {
			log.Printf("[CHARACTER_DATA] - Character Name: %s", name)
		} else if name, ok := data["name"].(string); ok {
			log.Printf("[CHARACTER_DATA] - Character Name (old format): %s", name)
		}
		if world, ok := data["world"].(string); ok {
			log.Printf("[CHARACTER_DATA] - World: %s", world)
		}

		// Log Glamourer data
		if glamourerData, ok := data["glamourer_data"].(string); ok {
			glamourerLength := len(glamourerData)
			log.Printf("[CHARACTER_DATA] - Glamourer Data: %d characters", glamourerLength)
			if glamourerLength > 0 {
				log.Printf("[CHARACTER_DATA] - Glamourer Preview: %s...", glamourerData[:min(50, glamourerLength)])
			}
		}

		// Log Penumbra meta data
		if penumbraMeta, ok := data["penumbra_meta"].(string); ok {
			metaLength := len(penumbraMeta)
			log.Printf("[CHARACTER_DATA] - Penumbra Meta: %d characters", metaLength)
			if metaLength > 0 {
				log.Printf("[CHARACTER_DATA] - Penumbra Meta Preview: %s...", penumbraMeta[:min(50, metaLength)])
			}
		}

		// Log Penumbra files data
		if penumbraFiles, ok := data["penumbra_files"].(map[string]interface{}); ok {
			fileCount := len(penumbraFiles)
			log.Printf("[CHARACTER_DATA] - Penumbra Files: %d files", fileCount)

			// Log details about each file
			totalSize := 0
			for fileName, fileData := range penumbraFiles {
				if fileBytes, ok := fileData.([]byte); ok {
					fileSize := len(fileBytes)
					totalSize += fileSize
					log.Printf("[CHARACTER_DATA]   - File: %s (%d bytes)", fileName, fileSize)
				}
			}
			log.Printf("[CHARACTER_DATA] - Total Penumbra Files Size: %d bytes (%.2f MB)", totalSize, float64(totalSize)/(1024*1024))
		}

		// Log Penumbra file data (compressed files)
		if penumbraFileData, ok := data["penumbra_file_data"].(map[string]interface{}); ok {
			compressedFileCount := len(penumbraFileData)
			log.Printf("[CHARACTER_DATA] - Penumbra File Data (Compressed): %d files", compressedFileCount)

			// Log details about each compressed file
			totalCompressedSize := 0
			for fileName, compressedData := range penumbraFileData {
				if fileBytes, ok := compressedData.([]byte); ok {
					fileSize := len(fileBytes)
					totalCompressedSize += fileSize
					log.Printf("[CHARACTER_DATA]   - Compressed File: %s (%d bytes)", fileName, fileSize)
				}
			}
			log.Printf("[CHARACTER_DATA] - Total Compressed Size: %d bytes (%.2f MB)", totalCompressedSize, float64(totalCompressedSize)/(1024*1024))
		}

		// Log Penumbra files (base64 encoded)
		if penumbraFiles, ok := data["penumbra_files"].(map[string]interface{}); ok {
			fileCount := len(penumbraFiles)
			log.Printf("[CHARACTER_DATA] - Penumbra Files (Base64): %d files", fileCount)

			// Log details about each base64 encoded file
			totalBase64Size := 0
			for fileName, base64Data := range penumbraFiles {
				if base64String, ok := base64Data.(string); ok {
					base64Size := len(base64String)
					totalBase64Size += base64Size

					// Calculate approximate original size (base64 is ~33% larger than binary)
					approximateOriginalSize := int(float64(base64Size) * 0.75)

					log.Printf("[CHARACTER_DATA]   - Base64 File: %s (%d base64 chars, ~%d bytes original)", fileName, base64Size, approximateOriginalSize)
				}
			}
			log.Printf("[CHARACTER_DATA] - Total Base64 Size: %d characters (%.2f MB)", totalBase64Size, float64(totalBase64Size)/(1024*1024))
		}
	} else {
		log.Printf("[CHARACTER_DATA] Warning: Received data is not a map for user %s", userName)
	}

	// Store the character data for this user
	client.GetServer().StoreUserData(userID, msg.Data)
	log.Printf("[CHARACTER_DATA] Stored character data for user %s (%s)", userName, userID)

	// Send acknowledgment to the sender
	response := models.Message{
		Type: "character_data_received",
		Data: map[string]interface{}{
			"received_at": time.Now().Unix(),
			"status":      "ok",
		},
	}
	client.SendMessage(response)
	log.Printf("[CHARACTER_DATA] Sent acknowledgment to user %s", userName)

	// Broadcast to other clients
	broadcastMsg := models.Message{
		Type: "character_data_broadcast",
		Data: msg.Data,
	}
	client.GetServer().BroadcastToOthers(client, broadcastMsg)
	log.Printf("[CHARACTER_DATA] Broadcasted character data from user %s to other clients", userName)
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
	log.Printf("=====================================")
	log.Printf("Starting Stellar Sync Server...")
	log.Printf("=====================================")

	// Get file server URL from environment variable, default to localhost
	fileServerURL := os.Getenv("FILE_SERVER_URL")
	if fileServerURL == "" {
		fileServerURL = "http://localhost:6200"
	}

	// Initialize file proxy to forward requests to file server
	fileProxy = proxy.NewFileProxy(fileServerURL)
	log.Printf("[STARTUP] File proxy initialized, forwarding to file server at %s", fileServerURL)

	// Create WebSocket server with message handler
	wsServer := websocket.NewServer(handleMessage)
	go wsServer.Start()
	log.Printf("[STARTUP] WebSocket server created and started")

	// Set up HTTP routes
	http.HandleFunc("/ws", wsServer.HandleWebSocket)
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/download", handleFileDownload)
	http.HandleFunc("/list", handleFileList)
	http.HandleFunc("/metadata/upload", handleFileMetadataUpload)
	http.HandleFunc("/metadata/download", handleFileMetadataDownload)
	log.Printf("[STARTUP] HTTP routes configured")

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
	log.Printf("[STARTUP] Main server configuration:")
	log.Printf("[STARTUP] - Port: %s", port)
	log.Printf("[STARTUP] - WebSocket endpoint: ws://localhost%s/ws", port)
	log.Printf("[STARTUP] - Status page: http://localhost%s/", port)
	log.Printf("[STARTUP] - Health check: http://localhost%s/health", port)
	log.Printf("[STARTUP] - File operations: Proxied to file server (port 6200)")
	log.Printf("[STARTUP] - WebSocket read limit: 100MB")
	log.Printf("=====================================")
	log.Printf("Main server started successfully!")
	log.Printf("File operations will be forwarded to file server at http://localhost:6200")
	log.Printf("Waiting for client connections...")
	log.Printf("=====================================")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("[FATAL] Server failed to start:", err)
	}
}
