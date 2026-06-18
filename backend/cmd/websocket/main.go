// @title HIS System WebSocket API
// @version 1.0
// @description Hospital Information System WebSocket Server
// @host localhost:8081
// @BasePath /api/ws
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"his-system/pkg/database"
	"his-system/pkg/middleware"
	"his-system/pkg/utils"

	"his-system/internal/reception/handlers"
	"his-system/internal/reception/infrastructure"

	commonLogger "github.com/thanhbvha/go-common/logger"
	commonRedis "github.com/thanhbvha/go-common/redis"
	commonWSFiber "github.com/thanhbvha/go-common/websocket/adapter/fiber"
	commonWSCore "github.com/thanhbvha/go-common/websocket/core"
)

func main() {
	// load env from file .env in development
	_ = godotenv.Load()

	// init logger
	logConfig := commonLogger.DefaultOptions()
	logConfig.Level = -4 // Debug (-4) | Info (0)
	appLogger := commonLogger.New(logConfig)
	commonLogger.SetDefault(appLogger)
	appLogger.InfoAsync("WebSocket Logger initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init database connection
	pgPool := database.MustNewPostgresPool(os.Getenv("POSTGRES_DSN"))
	defer pgPool.Close()
	appLogger.InfoAsync("PostgreSQL connected (WebSocket)", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// connect to redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	parts := strings.Split(redisAddr, ":")
	host := parts[0]
	port := 6379
	if len(parts) > 1 {
		port, _ = strconv.Atoi(parts[1])
	}
	rdb := commonRedis.MustConnect(context.Background(), commonRedis.Config{
		Host:     host,
		Port:     port,
		Password: "",
		DB:       0,
		Prefix:   os.Getenv("REDIS_PREFIX"),
	})
	commonRedis.SetDefault(rdb)
	appLogger.InfoAsync("Redis connected (WebSocket)", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init websocket manager
	wsManager := commonWSCore.GetGlobalManager()
	utils.SafeGo(func() { wsManager.Run() })
	appLogger.InfoAsync("WebSocket manager running", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init fiber app
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(middleware.RequestLogger(appLogger))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://wails.localhost:34115, http://wails.localhost, wails://wails",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Timestamp, X-Signature",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(middleware.Recover(appLogger))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Setup Reception Queue Handler for Custom WebSocket
	queueRepo := infrastructure.NewQueueRepositoryPG(pgPool)
	queueHandler := handlers.NewQueueHandler(queueRepo)
	customWsHandler := queueHandler.WSHandlerFactory()

	// Generic WebSocket Handler (if needed, or we just use Custom for everything)
	defaultWsHandler := commonWSFiber.NewHandler(commonWSFiber.Config{
		Authenticate: func(c *fiber.Ctx) (string, error) {
			token := c.Query("token")
			if token == "" {
				return "", fmt.Errorf("missing token")
			}
			return token, nil
		},
	})

	// Route mapping
	// Custom WS for Reception
	app.Get("/ws/queue", customWsHandler.HandleUpgrade)

	// Default WS for general usage
	app.Get("/ws", defaultWsHandler.HandleUpgrade)

	wsAPI := app.Group("/api/ws")
	wsAPI.Get("/stats", defaultWsHandler.HandleStats)
	wsAPI.All("/shard", defaultWsHandler.HandleShardManagement)

	// run server
	wsPort := os.Getenv("WS_PORT")
	if wsPort == "" {
		wsPort = "8081"
	}
	utils.SafeGo(func() {
		if err := app.Listen(":" + wsPort); err != nil {
			appLogger.ErrorAsync("Error starting WebSocket server", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.InfoAsync("Gracefully shutting down WebSocket server...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	if err := app.Shutdown(); err != nil {
		appLogger.ErrorAsync("WebSocket server shutdown failed", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	}
	appLogger.InfoAsync("WebSocket server exited properly", "dispatch_time", time.Now().Format(time.RFC3339Nano))
}
