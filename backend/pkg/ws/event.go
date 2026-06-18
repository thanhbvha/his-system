package ws

import (
	"encoding/json"

	"github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/websocket/core"
)

// WSEvent defines the payload format broadcasted to clients
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Event types
const (
	EventQueueUpdated   = "queue.updated"
	EventQueueCalled    = "queue.called"
	EventQueueCompleted = "queue.completed"
)

// BroadcastToAll is a helper to broadcast a JSON event to all connected clients globally
func BroadcastToAll(eventType string, payload interface{}) {
	manager := core.GetGlobalManager()
	event := WSEvent{Type: eventType, Payload: payload}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.ErrorAsync("Failed to marshal WS event for broadcast", "error", err, "eventType", eventType)
		return
	}

	manager.BroadcastToAll(eventBytes)
}
