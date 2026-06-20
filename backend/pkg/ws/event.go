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
	EventQueueSync      = "queue.sync"
	EventQueueCheckedIn = "queue.checked_in"
	EventQueueUpdated   = "queue.updated"
	EventQueueCalled    = "queue.called"
	EventQueueCompleted = "queue.completed"
	EventQueueSkipped   = "queue.skipped"
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

	// Tạm thời broadcast vào "default" shard của Manager.
	// API process chỉ khởi tạo mặc định "default" shard nên lệnh này sẽ đẩy sang Redis qua kênh shard:default
	manager.BroadcastToAll(eventBytes)
}
