package websocket

import "stellarsync-server/internal/models"

// ClientInterface defines the interface for WebSocket clients
type ClientInterface interface {
	SendMessage(msg models.Message)
}

// ServerInterface defines the interface for WebSocket server operations
type ServerInterface interface {
	BroadcastToOthers(sender ClientInterface, msg models.Message)
}
