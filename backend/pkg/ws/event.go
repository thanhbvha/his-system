package ws

import (
	"encoding/json"

	"github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/websocket/pubsub"
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

// BroadcastToRoom is a helper to broadcast a JSON event to a specific Room Shard
func BroadcastToRoom(roomID string, eventType string, payload interface{}) {
	event := WSEvent{Type: eventType, Payload: payload}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.ErrorAsync("Failed to marshal WS event for broadcast", "error", err, "eventType", eventType)
		return
	}

	pubsubManager := pubsub.GetGlobalPubSub()
	data := map[string]interface{}{
		"message": string(eventBytes),
	}

	shardID := "default"
	if roomID != "" {
		shardID = "room_" + roomID
	}
	
	// Broadcast to the specific room shard
	_ = pubsubManager.BroadcastMessage(shardID, data)
	
	// ALWAYS broadcast to global_reception shard so the Front Desk Receptionists 
	// can see all real-time updates across all clinic rooms.
	if shardID != "room_global_reception" {
		_ = pubsubManager.BroadcastMessage("room_global_reception", data)
	}
}
