package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"stellarsync-server/internal/models"
)

// Global file storage - in production, this would be a proper database
var fileStorage = make(map[string]models.FileMetadata)

// Global file metadata storage - for sharing file metadata between clients
var fileMetadataStorage = make(map[string]map[string]interface{})

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
	fileStorage[hash] = models.FileMetadata{
		Hash:         hash,
		Size:         written,
		ContentType:  header.Header.Get("Content-Type"),
		UploadTime:   time.Now().Unix(),
		FileName:     header.Filename,
		RelativePath: relativePath,
	}

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

	// Get hash from query parameter
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}

	// Check if file exists
	metadata, exists := fileStorage[hash]
	if !exists {
		log.Printf("[FILE_DOWNLOAD] File not found: %s", hash)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Get file path
	filePath := filepath.Join("./uploads", hash)

	// Check if file exists on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[FILE_DOWNLOAD] File not found on disk: %s", hash)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", metadata.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))

	// Serve file
	http.ServeFile(w, r, filePath)

	log.Printf("[FILE_DOWNLOAD] Successfully served file: %s", hash)
}

// handleFileList returns list of available files
func handleFileList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": fileStorage,
		"count": len(fileStorage),
	})
}

// handleFileMetadataUpload stores file metadata for a user
func handleFileMetadataUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		UserID       string                 `json:"user_id"`
		FileMetadata map[string]interface{} `json:"file_metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("[FILE_METADATA_UPLOAD] Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if requestData.UserID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	fileMetadataStorage[requestData.UserID] = requestData.FileMetadata
	log.Printf("[FILE_METADATA_UPLOAD] Stored file metadata for user %s: %d files", requestData.UserID, len(requestData.FileMetadata))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"files":  len(requestData.FileMetadata),
	})
}

// handleFileMetadataDownload retrieves file metadata for a user
func handleFileMetadataDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	metadata, exists := fileMetadataStorage[userID]
	if !exists {
		log.Printf("[FILE_METADATA_DOWNLOAD] No metadata found for user: %s", userID)
		http.Error(w, "No metadata found", http.StatusNotFound)
		return
	}

	log.Printf("[FILE_METADATA_DOWNLOAD] Retrieved file metadata for user %s: %d files", userID, len(metadata))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"files":   metadata,
		"count":   len(metadata),
	})
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "stellar-sync-fileserver",
	})
}

func main() {
	log.Printf("=====================================")
	log.Printf("Starting Stellar Sync File Server...")
	log.Printf("=====================================")

	// Set up HTTP routes
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
    <title>Stellar Sync File Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 10px; background: #e8f5e8; border: 1px solid #4caf50; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>Stellar Sync File Server</h1>
    <div class="status">
        <strong>Status:</strong> Running<br>
        <strong>Health Check:</strong> <a href="/health">/health</a>
    </div>
    <p>This server handles file operations for Stellar Sync clients.</p>
    <h3>Available Endpoints:</h3>
    <ul>
        <li><code>/upload</code> - File upload</li>
        <li><code>/download?hash=...</code> - File download</li>
        <li><code>/list</code> - List all files</li>
        <li><code>/metadata/upload</code> - Upload file metadata</li>
        <li><code>/metadata/download?user_id=...</code> - Download file metadata</li>
    </ul>
</body>
</html>
`)
	})

	port := ":6200"
	log.Printf("[STARTUP] File server configuration:")
	log.Printf("[STARTUP] - Port: %s", port)
	log.Printf("[STARTUP] - Status page: http://localhost%s/", port)
	log.Printf("[STARTUP] - Health check: http://localhost%s/health", port)
	log.Printf("[STARTUP] - File upload: http://localhost%s/upload", port)
	log.Printf("[STARTUP] - File download: http://localhost%s/download", port)
	log.Printf("[STARTUP] - File list: http://localhost%s/list", port)
	log.Printf("[STARTUP] - Metadata upload: http://localhost%s/metadata/upload", port)
	log.Printf("[STARTUP] - Metadata download: http://localhost%s/metadata/download", port)
	log.Printf("=====================================")
	log.Printf("File server started successfully!")
	log.Printf("Waiting for file operations...")
	log.Printf("=====================================")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("[FATAL] File server failed to start:", err)
	}
}
