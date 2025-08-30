package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"stellarsync-server/internal/models"
	"stellarsync-server/internal/websocket"
)

// Global file storage - in production, this would be a proper database
var fileStorage = make(map[string]models.FileMetadata)
var fileStorageMutex sync.RWMutex

// Global file metadata storage - for sharing file metadata between clients
var fileMetadataStorage = make(map[string]map[string]interface{})
var fileMetadataMutex sync.RWMutex

// WebSocket server instance
var wsServer *websocket.Server

// Message handler function for WebSocket server
func handleWebSocketMessage(client *websocket.Client, msg models.Message) {
	log.Printf("Received message: %+v", msg)

	switch msg.Type {
	case "connect":
		handleConnect(client, msg)
	case "character_data":
		handleCharacterData(client, msg)
	case "request_users":
		handleRequestUsers(client, msg)
	case "request_user_data":
		handleRequestUserData(client, msg)
	default:
		handleUnknownMessage(client, msg)
	}
}

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
}

func handleCharacterData(client *websocket.Client, msg models.Message) {
	log.Printf("[CHARACTER_DATA] Received character data from user")
	// Store character data for the user
	wsServer.StoreUserData(client.GetUserID(), msg.Data)

	// Broadcast to all clients that new data is available
	wsServer.BroadcastToOthers(client, models.Message{
		Type: "character_data_received",
		Data: map[string]interface{}{
			"user_id": client.GetUserID(),
			"time":    time.Now().Unix(),
		},
	})
}

func handleRequestUsers(client *websocket.Client, msg models.Message) {
	log.Printf("[REQUEST_USERS] User requesting online users list")

	onlineUsers := wsServer.GetOnlineUsers()
	response := models.Message{
		Type: "users_list",
		Data: onlineUsers,
	}
	client.SendMessage(response)
}

func handleRequestUserData(client *websocket.Client, msg models.Message) {
	log.Printf("[REQUEST_USER_DATA] User requesting specific user data")

	// Extract target user ID
	var targetUserID string
	if data, ok := msg.Data.(map[string]interface{}); ok {
		if id, ok := data["user_id"].(string); ok {
			targetUserID = id
		}
	}

	if targetUserID == "" {
		log.Printf("[REQUEST_USER_DATA] No target user ID provided")
		return
	}

	// Get stored character data for the target user
	userData, exists := wsServer.GetUserData(targetUserID)
	if exists {
		// Find the user name from online users
		var userName string
		onlineUsers := wsServer.GetOnlineUsers()
		for _, user := range onlineUsers {
			if user.ID == targetUserID {
				userName = user.Name
				break
			}
		}

		response := models.Message{
			Type: "user_character_data",
			Data: map[string]interface{}{
				"data":                  userData,
				"source_character_name": userName,
			},
		}
		client.SendMessage(response)
		log.Printf("[REQUEST_USER_DATA] Sent character data for user %s", targetUserID)
	} else {
		log.Printf("[REQUEST_USER_DATA] No character data found for user %s", targetUserID)
	}
}

func handleUnknownMessage(client *websocket.Client, msg models.Message) {
	log.Printf("[UNKNOWN] Unknown message type: %s", msg.Type)
}

func main() {
	// Get port from environment variable (Koyeb sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "6000" // Default fallback
	}

	// Initialize WebSocket server
	wsServer = websocket.NewServer(handleWebSocketMessage)
	go wsServer.Start()

	// Set up HTTP routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/ws", handleWebSocket)

	// File server routes
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/download/", handleFileDownload)
	http.HandleFunc("/files", handleFileList)
	http.HandleFunc("/metadata/upload", handleFileMetadataUpload)
	http.HandleFunc("/metadata/download/", handleFileMetadataDownload)

	// Create uploads directory
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Failed to create uploads directory: %v", err)
	}

	log.Printf("Starting Stellar Sync Server on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("File upload endpoint: http://localhost:%s/upload", port)
	log.Printf("File download endpoint: http://localhost:%s/download/", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "running",
		"service": "Stellar Sync Server",
		"time":    time.Now().Unix(),
		"endpoints": map[string]string{
			"websocket":         "/ws",
			"upload":            "/upload",
			"download":          "/download/",
			"files":             "/files",
			"metadata_upload":   "/metadata/upload",
			"metadata_download": "/metadata/download/",
		},
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsServer.HandleWebSocket(w, r)
}

// handleFileUpload handles file uploads from clients
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("[FILE_UPLOAD] Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("[FILE_UPLOAD] Failed to get file: %v", err)
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get metadata from form
	hash := r.FormValue("hash")
	relativePath := r.FormValue("relative_path")

	if hash == "" || relativePath == "" {
		log.Printf("[FILE_UPLOAD] Missing required metadata")
		http.Error(w, "Missing required metadata", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("[FILE_UPLOAD] Failed to create uploads directory: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create file path
	filePath := filepath.Join(uploadsDir, hash)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("[FILE_UPLOAD] Failed to create file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("[FILE_UPLOAD] Failed to save file: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Store metadata
	fileStorageMutex.Lock()
	fileStorage[hash] = models.FileMetadata{
		Hash:         hash,
		Size:         written,
		ContentType:  header.Header.Get("Content-Type"),
		UploadTime:   time.Now().Unix(),
		FileName:     header.Filename,
		RelativePath: relativePath,
	}
	fileStorageMutex.Unlock()

	log.Printf("[FILE_UPLOAD] Successfully uploaded file: %s (%d bytes)", hash, written)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"hash":   hash,
		"size":   written,
	})
}

// handleFileDownload handles file downloads
func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hash from URL path
	hash := r.URL.Path[len("/download/"):]
	if hash == "" {
		http.Error(w, "Hash parameter required", http.StatusBadRequest)
		return
	}

	// Check if file exists
	fileStorageMutex.RLock()
	_, exists := fileStorage[hash]
	fileStorageMutex.RUnlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Check if file exists on disk
	filePath := filepath.Join("./uploads", hash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
	log.Printf("[FILE_DOWNLOAD] Served file: %s", hash)
}

// handleFileList returns a list of available files
func handleFileList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileStorageMutex.RLock()
	files := make([]models.FileMetadata, 0, len(fileStorage))
	for _, metadata := range fileStorage {
		files = append(files, metadata)
	}
	fileStorageMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

// handleFileMetadataUpload handles file metadata uploads
func handleFileMetadataUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var metadata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		log.Printf("[METADATA_UPLOAD] Failed to decode JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract user ID and file metadata
	userID, ok := metadata["user_id"].(string)
	if !ok {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	fileMetadata, ok := metadata["file_metadata"].(map[string]interface{})
	if !ok {
		http.Error(w, "file_metadata required", http.StatusBadRequest)
		return
	}

	// Store metadata
	fileMetadataMutex.Lock()
	if fileMetadataStorage[userID] == nil {
		fileMetadataStorage[userID] = make(map[string]interface{})
	}
	fileMetadataStorage[userID] = fileMetadata
	fileMetadataMutex.Unlock()

	log.Printf("[METADATA_UPLOAD] Stored metadata for user: %s", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"user_id": userID,
	})
}

// handleFileMetadataDownload handles file metadata downloads
func handleFileMetadataDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := r.URL.Path[len("/metadata/download/"):]
	if userID == "" {
		http.Error(w, "User ID parameter required", http.StatusBadRequest)
		return
	}

	// Get metadata
	fileMetadataMutex.RLock()
	metadata, exists := fileMetadataStorage[userID]
	fileMetadataMutex.RUnlock()

	if !exists {
		http.Error(w, "Metadata not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       userID,
		"file_metadata": metadata,
	})
}
