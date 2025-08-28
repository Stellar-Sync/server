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
	GlamourerData string                 `json:"glamourer_data"`
	PenumbraMeta  string                 `json:"penumbra_meta"`
	PenumbraFiles map[string]interface{} `json:"penumbra_files"`
	Timestamp     int64                  `json:"timestamp"`
}
