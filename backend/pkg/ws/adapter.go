package ws

import (


	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/thanhbvha/go-common/logger"
	libadapter "github.com/thanhbvha/go-common/websocket/adapter/fiber"
	"github.com/thanhbvha/go-common/websocket/core"
	"github.com/thanhbvha/go-common/websocket/limiter"
)

// CustomWSHandler integrates go-common/websocket with fiber and provides an OnConnect hook.
type CustomWSHandler struct {
	// Authenticate resolves user identifiers from the Fiber context.
	Authenticate func(c *fiber.Ctx) (string, error)
	// OnConnect is called immediately after the connection is upgraded and registered,
	// allowing us to send initial state data before blocking in the read pump.
	OnConnect func(userID string, sendJSON func(interface{}) bool)

	RateLimiter       *limiter.RateLimiter
	ConnectionLimiter *limiter.ConnectionLimiter
}

// NewCustomWSHandler creates a new handler with sensible defaults.
func NewCustomWSHandler() *CustomWSHandler {
	return &CustomWSHandler{
		RateLimiter:       limiter.GetGlobalRateLimiter(),
		ConnectionLimiter: limiter.GetGlobalConnectionLimiter(),
	}
}

// HandleUpgrade handles the Fiber WebSocket connection upgrade process.
func (h *CustomWSHandler) HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		logger.WarnAsync("Received non-websocket upgrade HTTP request")
		return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{"error": "upgrade required"})
	}

	userID, err := h.Authenticate(c)
	if err != nil || userID == "" {
		logger.WarnAsync("WebSocket connection upgrade unauthorized", "error", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	clientIP := c.IP()
	requestID := ""
	if reqIDVal := c.Locals("requestid"); reqIDVal != nil {
		if reqIDStr, ok := reqIDVal.(string); ok {
			requestID = reqIDStr
		}
	}

	return websocket.New(func(conn *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorAsync("Panic in CustomWSHandler connection loop", "error", r, "userID", userID)
			}
		}()

		if conn == nil {
			return
		}

		if !h.RateLimiter.Allow() {
			logger.WarnAsync("Connection upgrade rejected due to rate limiting", "clientIP", clientIP, "userID", userID)
			_ = conn.Close()
			return
		}

		if !h.ConnectionLimiter.CanConnect(clientIP) {
			logger.WarnAsync("Connection upgrade rejected due to key connection limit", "clientIP", clientIP, "userID", userID)
			_ = conn.Close()
			return
		}

		if !h.ConnectionLimiter.AddConnection(clientIP) {
			_ = conn.Close()
			return
		}

		defer h.ConnectionLimiter.RemoveConnection(clientIP)

		manager := core.GetGlobalManager()
		shardID := manager.GetShardID(userID)

		adapterConn := libadapter.NewConnAdapter(conn)

		// Manual HandleConnection inline so we can inject OnConnect
		if !manager.CanAcceptConnection() {
			logger.WarnAsync("Manager connection ceiling reached, rejecting connection", "total", manager.GetTotalConnections())
			_ = adapterConn.Close()
			return
		}

		shard := manager.GetOrCreateShard(shardID)
		if !shard.CanAcceptConnection() {
			logger.WarnAsync("Shard connection ceiling reached, rejecting connection", "shardID", shardID)
			_ = adapterConn.Close()
			return
		}

		// The NodeID is exposed via pubsub. We can use a dummy or just use manager's methods if possible.
		// Since we can't easily get NodeID from Manager (it's private if no getter is exported),
		// we'll just pass an empty string or standard "node" since core handles internal routing.
		// Wait, core.NewConnection requires nodeID. Let's look at libadapter:
		// manager.HandleConnection handles this all automatically!
		// BUT HandleConnection blocks inside readPump.
		// Let's run HandleConnection in a goroutine, but wait, HandleConnection blocking prevents websocket.New from returning?
		// No, websocket.New expects the func to block. If the func returns, the connection closes!
		// So we CANNOT run HandleConnection in a goroutine without keeping the func alive.
		// But if we run manager.HandleConnection, how do we trigger OnConnect?
		
		// Actually, what we can do is call OnConnect in a separate goroutine immediately before HandleConnection!
		// But how do we pass `sendJSON`? The manager doesn't give us the `core.Connection`.
		// But we can just use `manager.BroadcastMessage(shardID, eventBytes)`? No, that broadcasts to the whole shard.
		
		// Let's create an adapter wrapper that sends the initial message immediately upon WritePump start?
		// Actually, if we just call OnConnect AFTER manager.HandleConnection, it's too late because it blocks.
		// If we call it BEFORE manager.HandleConnection, we only have `conn *websocket.Conn`. We can just write directly to `conn`!
		
		if h.OnConnect != nil {
			h.OnConnect(userID, func(v interface{}) bool {
				err := conn.WriteJSON(v)
				if err != nil {
					logger.ErrorAsync("Failed to send initial JSON on connect", "error", err, "userID", userID)
					return false
				}
				return true
			})
		}

		// Now hand over to manager which will block and handle read/write pumps
		if err := manager.HandleConnection(adapterConn, shardID, userID, clientIP, requestID); err != nil {
			logger.ErrorAsync("Failed to handle WebSocket connection in manager", "error", err, "userID", userID)
		}
	})(c)
}
