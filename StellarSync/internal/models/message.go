package models

// Message represents a JSON message sent between client and server
type Message struct {
	Type   string      `json:"type"`
	Client string      `json:"client,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// CharacterData represents character data sent from clients
type CharacterData struct {
	GlamourerData    string                 `json:"glamourer_data"`
	PenumbraMeta     string                 `json:"penumbra_meta"`
	PenumbraFiles    map[string]interface{} `json:"penumbra_files"`
	PenumbraFileData map[string][]byte      `json:"penumbra_file_data"` // Compressed mod files
	Timestamp        int64                  `json:"timestamp"`
	Zone             string                 `json:"zone,omitempty"` // Current zone/area the character is in
}

// FileMetadata represents file information stored on the server
type FileMetadata struct {
	Hash         string `json:"hash"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	UploadTime   int64  `json:"upload_time"`
	FileName     string `json:"file_name"`
	RelativePath string `json:"relative_path"`
}

// UserInfo represents information about a connected user
type UserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Zone     string `json:"zone,omitempty"` // Current zone/area the user is in
	Online   bool   `json:"online"`
	LastSeen int64  `json:"last_seen"`
}
