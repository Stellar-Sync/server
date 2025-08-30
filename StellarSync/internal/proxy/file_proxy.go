package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// FileProxy handles forwarding file requests to the file server
type FileProxy struct {
	fileServerURL string
	httpClient    *http.Client
}

// NewFileProxy creates a new file proxy instance
func NewFileProxy(fileServerURL string) *FileProxy {
	return &FileProxy{
		fileServerURL: fileServerURL,
		httpClient:    &http.Client{},
	}
}

// ForwardRequest forwards an HTTP request to the file server
func (fp *FileProxy) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	// Build the target URL
	targetURL := fp.fileServerURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	log.Printf("[PROXY] Forwarding %s %s to %s", r.Method, r.URL.Path, targetURL)

	// Create the request to the file server
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("[PROXY] Failed to create request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Set the host header to the file server
	parsedURL, _ := url.Parse(fp.fileServerURL)
	req.Host = parsedURL.Host

	// Make the request to the file server
	resp, err := fp.httpClient.Do(req)
	if err != nil {
		log.Printf("[PROXY] Failed to forward request: %v", err)
		http.Error(w, "Failed to connect to file server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the response status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[PROXY] Failed to copy response body: %v", err)
	}

	log.Printf("[PROXY] Successfully forwarded request to file server")
}

// UploadFileMetadata forwards file metadata upload to the file server
func (fp *FileProxy) UploadFileMetadata(userID string, fileMetadata map[string]interface{}) error {
	requestData := map[string]interface{}{
		"user_id":       userID,
		"file_metadata": fileMetadata,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	targetURL := fp.fileServerURL + "/metadata/upload"
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := fp.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to forward metadata upload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("file server returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[PROXY] Successfully uploaded file metadata for user %s", userID)
	return nil
}

// DownloadFileMetadata forwards file metadata download from the file server
func (fp *FileProxy) DownloadFileMetadata(userID string) (map[string]interface{}, error) {
	targetURL := fmt.Sprintf("%s/metadata/download?user_id=%s", fp.fileServerURL, userID)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := fp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to forward metadata download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file server returned status %d: %s", resp.StatusCode, string(body))
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract the files from the response
	if files, ok := responseData["files"].(map[string]interface{}); ok {
		log.Printf("[PROXY] Successfully downloaded file metadata for user %s", userID)
		return files, nil
	}

	return nil, fmt.Errorf("invalid response format from file server")
}

// UploadFile forwards a file upload to the file server
func (fp *FileProxy) UploadFile(filePath, hash, relativePath string) error {
	// This would need to be implemented to handle multipart form data
	// For now, we'll just forward the request as-is
	return fmt.Errorf("file upload forwarding not yet implemented")
}

// DownloadFile forwards a file download from the file server
func (fp *FileProxy) DownloadFile(hash string) ([]byte, error) {
	targetURL := fmt.Sprintf("%s/download?hash=%s", fp.fileServerURL, hash)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := fp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to forward file download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file server returned status %d: %s", resp.StatusCode, string(body))
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %v", err)
	}

	log.Printf("[PROXY] Successfully downloaded file %s (%d bytes)", hash, len(fileData))
	return fileData, nil
}
