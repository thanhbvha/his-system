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
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber/v2"

	appointmentMod "his-system/internal/appointment/bootstrap"
	adminMod "his-system/internal/identity/bootstrap/admin"
	authMod "his-system/internal/identity/bootstrap/auth"
	patientMod "his-system/internal/patient/bootstrap"
	publicMod "his-system/internal/public/bootstrap"
	"his-system/internal/system/handlers"

	"his-system/pkg/crypto"
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
	// load env from file .env in development
	_ = godotenv.Load()

	// init logger
	logConfig := commonLogger.DefaultOptions()
	logConfig.Level = -4 // Debug (-4) | Info (0)
	appLogger := commonLogger.New(logConfig)
	commonLogger.SetDefault(appLogger)
	appLogger.InfoAsync("Logger initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init tracer (optional)
	shutdownTracer, err := telemetry.InitTracer("his-system", "localhost:4317")
	if err != nil {
		appLogger.ErrorAsync("Failed to init tracer", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	} else {
		appLogger.InfoAsync("OpenTelemetry tracer initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				appLogger.ErrorAsync("Failed to shutdown tracer", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
			}
		}()
	}

	// init database connection
	pgPool := database.MustNewPostgresPool(os.Getenv("POSTGRES_DSN"))
	defer pgPool.Close()
	appLogger.InfoAsync("PostgreSQL connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// connect to mongodb
	mongoClient := database.MustNewMongoClient(os.Getenv("MONGO_URI"))
	appLogger.InfoAsync("MongoDB connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

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
	appLogger.InfoAsync("Redis connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init queue
	q := commonQueue.New(rdb, commonQueue.Config{})
	appLogger.InfoAsync("Queue initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init websocket manager
	wsManager := commonWSCore.GetGlobalManager()
	utils.SafeGo(func() { wsManager.Run() })
	appLogger.InfoAsync("WebSocket manager running", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// init minio client
	minioClient, err := storage.NewMinioClient(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		false,
	)
	if err != nil {
		appLogger.ErrorAsync("Failed to init MinIO", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	} else {
		appLogger.InfoAsync("MinIO client initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	}

	_ = mongoClient
	_ = minioClient

	// init jwt and field cipher
	signKey := []byte(os.Getenv("JWT_SIGNING_KEY"))
	encKey := []byte(os.Getenv("JWT_ENCRYPTION_KEY"))

	fieldEncKey, _ := base64.StdEncoding.DecodeString(os.Getenv("FIELD_ENCRYPTION_KEY"))
	fieldHMACKey, _ := base64.StdEncoding.DecodeString(os.Getenv("FIELD_HMAC_KEY"))
	cipher, err := crypto.NewFieldCipher(fieldEncKey, fieldHMACKey)
	if err != nil {
		appLogger.ErrorAsync("Failed to init field cipher", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	}

	// init fiber app
	app := fiber.New()

	// register prometheus middleware
	prometheus := fiberprometheus.New("his-system")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(requestid.New())
	app.Use(middleware.RequestLogger(appLogger))
	app.Use(otelfiber.Middleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://wails.localhost:34115, http://wails.localhost, wails://wails",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Timestamp, X-Signature",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:          100,
		Expiration:   1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return response.Fail(c, &appErrors.AppError{Code: "TOO_MANY_REQUESTS", Status: 429})
		},
	}))
	app.Use(middleware.Recover(appLogger))

	// init health handler
	healthHandler := handlers.NewHealthHandler(pgPool, mongoClient, rdb)
	app.Get("/health", healthHandler.CheckLiveness)
	app.Get("/ready", healthHandler.CheckReadiness)
	app.Static("/static", "./public")

	// jwt auth middleware
	jwtAuth := middleware.JWTAuth(signKey, encKey)
	api := app.Group("/api/v1")

	// /api/v1/auth
	authMod.NewModule(authMod.ModuleDeps{
		PgPool:  pgPool,
		Rdb:     rdb,
		Queue:   q,
		Cipher:  cipher,
		SignKey: signKey,
		EncKey:  encKey,
	}).Register(api.Group("/auth"))

	// /api/v1/admin - Desktop only, Hardware Signature required
	adminMod.NewModule(adminMod.ModuleDeps{
		PgPool: pgPool,
		Rdb:    rdb,
		Cipher: cipher,
	}).Register(api.Group("/admin",
		jwtAuth,
		middleware.RequireRole("admin"),
		adminMod.DeviceMiddleware(pgPool),
	))

	// /api/v1/patients
	patientMod.NewModule(patientMod.ModuleDeps{
		PgPool:  pgPool,
		Rdb:     rdb,
		Cipher:  cipher,
		SignKey: signKey,
		EncKey:  encKey,
	}).Register(api.Group("/patients", jwtAuth))

	// /api/v1/appointments
	appointmentMod.NewModule(appointmentMod.ModuleDeps{
		PgPool: pgPool,
		Rdb:    rdb,
	}).Register(api.Group("/appointments", jwtAuth))

	// /api/v1/public
	publicMod.NewModule(publicMod.ModuleDeps{}).Register(api.Group("/public"))

	// docs
	docsHandler := handlers.NewDocsHandler()
	app.Get("/docs/tool", docsHandler.ServeSwaggerUI)
	app.Get("/docs", docsHandler.ServeReDoc)
	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// websocket
	wsHandler := commonWSFiber.NewHandler(commonWSFiber.Config{
		Authenticate: func(c *fiber.Ctx) (string, error) {
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

	// run server
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	utils.SafeGo(func() {
		if err := app.Listen(":" + apiPort); err != nil {
			appLogger.ErrorAsync("Error starting server", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.InfoAsync("Gracefully shutting down server...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	if err := app.Shutdown(); err != nil {
		appLogger.ErrorAsync("Server shutdown failed", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	}
	appLogger.InfoAsync("Server exited properly", "dispatch_time", time.Now().Format(time.RFC3339Nano))
}
