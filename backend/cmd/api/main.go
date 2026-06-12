// @title HIS System API
// @version 1.0
// @description Hospital Information System API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/ansrivas/fiberprometheus/v2"

	"his-system/internal/api/handlers"
	"his-system/pkg/database"
	appErrors "his-system/pkg/errors"
	"his-system/pkg/middleware"
	"his-system/pkg/response"
	"his-system/pkg/storage"
	"his-system/pkg/telemetry"
	"his-system/pkg/utils"

	commonLogger "github.com/thanhbvha/go-common/logger"
	commonQueue "github.com/thanhbvha/go-common/queue"
	commonRedis "github.com/thanhbvha/go-common/redis"
	commonWSFiber "github.com/thanhbvha/go-common/websocket/adapter/fiber"
	commonWSCore "github.com/thanhbvha/go-common/websocket/core"
)

func main() {
	// 1. Load env
	_ = godotenv.Load()

	// 2. Init logger
	logConfig := commonLogger.DefaultOptions()
	logConfig.Level = -4 // Debug or 0 for Info
	appLogger := commonLogger.New(logConfig)
	commonLogger.SetDefault(appLogger)
	appLogger.InfoAsync("Logger initialized")

	// 2.5 Init Tracer
	shutdownTracer, err := telemetry.InitTracer("his-system", "localhost:4317")
	if err != nil {
		appLogger.ErrorAsync("Failed to init tracer", "error", err.Error())
	} else {
		appLogger.InfoAsync("OpenTelemetry tracer initialized")
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				appLogger.ErrorAsync("Failed to shutdown tracer", "error", err.Error())
			}
		}()
	}

	// 3. Init DB
	pgPool := database.MustNewPostgresPool(os.Getenv("POSTGRES_DSN"))
	defer pgPool.Close()
	appLogger.InfoAsync("PostgreSQL connected")

	mongoClient := database.MustNewMongoClient(os.Getenv("MONGO_URI"))
	appLogger.InfoAsync("MongoDB connected")

	// 4. Init Redis
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
		Password: "", // or load from env
		DB:       0,
	})
	appLogger.InfoAsync("Redis connected")

	// 5. Init Queue
	q := commonQueue.New(rdb, commonQueue.Config{})
	appLogger.InfoAsync("Queue initialized")

	// 6. Init WebSocket Manager
	wsManager := commonWSCore.GetGlobalManager()
	utils.SafeGo(func() {
		wsManager.Run()
	})
	appLogger.InfoAsync("WebSocket manager running")

	// 7. Init Storage
	minioClient, err := storage.NewMinioClient(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		false,
	)
	if err != nil {
		appLogger.ErrorAsync("Failed to init MinIO", "error", err.Error())
	} else {
		appLogger.InfoAsync("MinIO client initialized")
	}

	_ = mongoClient
	_ = rdb
	_ = q
	_ = minioClient

	// 8. Bootstrap Fiber app
	app := fiber.New()

	prometheus := fiberprometheus.New("his-system")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// 9. Register middleware
	app.Use(requestid.New())
	app.Use(middleware.RequestLogger(appLogger))
	app.Use(otelfiber.Middleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Fail(c, &appErrors.AppError{Code: "TOO_MANY_REQUESTS", Status: 429})
		},
	}))
	app.Use(middleware.Recover(appLogger))

	// 10. Register routes
	healthHandler := handlers.NewHealthHandler(pgPool, mongoClient, rdb)
	app.Get("/health", healthHandler.CheckLiveness)
	app.Get("/ready", healthHandler.CheckReadiness)

	app.Static("/static", "./public")

	docsHandler := handlers.NewDocsHandler()
	app.Get("/docs/tool", docsHandler.ServeSwaggerUI)
	app.Get("/docs", docsHandler.ServeReDoc)
	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// WebSocket routes
	wsHandler := commonWSFiber.NewHandler(commonWSFiber.Config{
		Authenticate: func(c *fiber.Ctx) (string, error) {
			// TODO: Implement JWT auth later (Step 5)
			token := c.Query("token")
			if token == "" {
				return "", fmt.Errorf("missing token")
			}
			return token, nil
		},
	})
	app.Get("/ws", wsHandler.HandleUpgrade)
	
	wsAPI := app.Group("/api/ws")
	wsAPI.Get("/stats", wsHandler.HandleStats)
	wsAPI.All("/shard", wsHandler.HandleShardManagement)

	// 11. Start server with graceful shutdown
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	utils.SafeGo(func() {
		if err := app.Listen(":" + apiPort); err != nil {
			appLogger.ErrorAsync("Error starting server", "error", err.Error())
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.InfoAsync("Gracefully shutting down server...")
	if err := app.Shutdown(); err != nil {
		appLogger.ErrorAsync("Server shutdown failed", "error", err.Error())
	}
	appLogger.InfoAsync("Server exited properly")
}
