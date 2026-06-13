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

	adminHandlers "his-system/internal/api/admin"
	authHandlers "his-system/internal/api/auth"
	"his-system/internal/api/handlers"
	"his-system/internal/identity/application/command"
	"his-system/internal/identity/application/query"
	identityInfra "his-system/internal/identity/infrastructure"
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
	// 1. Load env
	_ = godotenv.Load()

	// 2. Init logger
	logConfig := commonLogger.DefaultOptions()
	logConfig.Level = -4 // Debug or 0 for Info
	appLogger := commonLogger.New(logConfig)
	commonLogger.SetDefault(appLogger)
	appLogger.InfoAsync("Logger initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// 2.5 Init Tracer
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

	// 3. Init DB
	pgPool := database.MustNewPostgresPool(os.Getenv("POSTGRES_DSN"))
	defer pgPool.Close()
	appLogger.InfoAsync("PostgreSQL connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	mongoClient := database.MustNewMongoClient(os.Getenv("MONGO_URI"))
	appLogger.InfoAsync("MongoDB connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

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
		Prefix:   os.Getenv("REDIS_PREFIX"),
	})
	commonRedis.SetDefault(rdb)
	appLogger.InfoAsync("Redis connected", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// 5. Init Queue
	q := commonQueue.New(rdb, commonQueue.Config{})
	appLogger.InfoAsync("Queue initialized", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// 6. Init WebSocket Manager
	wsManager := commonWSCore.GetGlobalManager()
	utils.SafeGo(func() {
		wsManager.Run()
	})
	appLogger.InfoAsync("WebSocket manager running", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// 7. Init Storage
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
		AllowOrigins:     "http://localhost:5173, http://wails.localhost:34115, http://wails.localhost, wails://wails",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Timestamp, X-Signature",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
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

	// Initialize Identity Infrastructure
	userRepo := identityInfra.NewUserRepositoryPG(pgPool)
	roleRepo := identityInfra.NewRoleRepositoryPG(pgPool)
	deviceRepo := identityInfra.NewDeviceRepositoryPG(pgPool)
	mfaRepo := identityInfra.NewMFARepositoryPG(pgPool)
	patientRepo := identityInfra.NewPatientRepositoryPG(pgPool)
	deptRepo := identityInfra.NewDepartmentRepositoryPG(pgPool)

	// Auth Keys
	signKey := []byte(os.Getenv("JWT_SIGNING_KEY"))
	encKey := []byte(os.Getenv("JWT_ENCRYPTION_KEY"))

	// Field Cipher
	fieldEncKey, _ := base64.StdEncoding.DecodeString(os.Getenv("FIELD_ENCRYPTION_KEY"))
	fieldHMACKey, _ := base64.StdEncoding.DecodeString(os.Getenv("FIELD_HMAC_KEY"))
	cipher, err := crypto.NewFieldCipher(fieldEncKey, fieldHMACKey)
	if err != nil {
		appLogger.ErrorAsync("Failed to init field cipher", "error", err.Error(), "dispatch_time", time.Now().Format(time.RFC3339Nano))
	}

	// Initialize Identity Commands
	initLoginCmd := command.NewInitLoginHandler(userRepo, rdb)
	completeLoginCmd := command.NewCompleteLoginHandler(userRepo, deviceRepo, roleRepo, rdb, signKey, encKey)
	refreshTokenCmd := command.NewRefreshTokenHandler(userRepo, roleRepo, rdb, signKey, encKey)
	logoutCmd := command.NewLogoutHandler(rdb)
	setupMFACmd := command.NewSetupMFAHandler(mfaRepo, encKey)
	verifyMFACmd := command.NewVerifyMFAHandler(mfaRepo, userRepo, rdb, encKey)

	// Initialize Web Auth Commands
	sendOTPCmd := command.NewSendOTPHandler(rdb, q, cipher)
	verifyOTPCmd := command.NewVerifyOTPHandler(patientRepo, rdb, cipher, signKey, encKey)
	registerPatientCmd := command.NewRegisterPatientHandler(userRepo, patientRepo, rdb, cipher, signKey, encKey)
	refreshWebCmd := command.NewRefreshWebHandler(patientRepo, rdb, signKey, encKey)
	logoutWebCmd := command.NewLogoutWebHandler(rdb)

	// Initialize Admin Commands & Queries
	getUserByIDCmd := query.NewGetUserByIDHandler(userRepo, roleRepo, cipher)
	listUsersCmd := query.NewListUsersHandler(userRepo, roleRepo, cipher)
	getRolePermsCmd := query.NewGetRolePermissionsHandler(roleRepo)
	listPermsCmd := query.NewListPermissionsHandler(roleRepo)

	createStaffCmd := command.NewCreateStaffHandler(userRepo, cipher)
	deactivateCmd := command.NewDeactivateUserHandler(userRepo, deviceRepo)
	assignRolesCmd := command.NewAssignUserRolesHandler(userRepo)
	updateRolePermsCmd := command.NewUpdateRolePermissionsHandler(roleRepo, rdb)
	createDeptCmd := command.NewCreateDepartmentHandler(deptRepo)

	// Initialize Handlers
	adminUserHandler := adminHandlers.NewAdminUserHandler(getUserByIDCmd, listUsersCmd, createStaffCmd, deactivateCmd, assignRolesCmd)
	adminRoleHandler := adminHandlers.NewAdminRoleHandler(roleRepo, getRolePermsCmd, updateRolePermsCmd, listPermsCmd)
	adminDeptHandler := adminHandlers.NewAdminDepartmentHandler(deptRepo, createDeptCmd)

	// Initialize Auth Handlers
	desktopAuthHandler := authHandlers.NewDesktopAuthHandler(
		initLoginCmd,
		completeLoginCmd,
		refreshTokenCmd,
		logoutCmd,
		setupMFACmd,
		verifyMFACmd,
	)

	webAuthHandler := authHandlers.NewWebAuthHandler(
		sendOTPCmd,
		verifyOTPCmd,
		registerPatientCmd,
		refreshWebCmd,
		logoutWebCmd,
	)

	// Auth Routes
	authGrp := app.Group("/api/v1/auth")
	authGrp.Use(middleware.AuthRateLimit(rdb)) // 5 req/min per IP

	// Desktop - Hardware-bound login
	authGrp.Post("/login/init", desktopAuthHandler.InitLogin)
	authGrp.Post("/login/complete", desktopAuthHandler.CompleteLogin)
	authGrp.Post("/mfa/verify", desktopAuthHandler.VerifyMFA)
	// mfa setup route is moved down to apply JWTAuth middleware
	authGrp.Post("/refresh", desktopAuthHandler.RefreshToken)
	authGrp.Post("/logout", desktopAuthHandler.Logout)

	// Web - Patient OTP flow
	authGrp.Post("/otp/send", webAuthHandler.SendOTP)
	authGrp.Post("/otp/verify", webAuthHandler.VerifyOTP)
	authGrp.Post("/register", webAuthHandler.Register)
	// Refresh & Logout for Web are deliberately on root /refresh and /logout or specific?
	// Let's mount them at /web/refresh to avoid clash with desktop, or use the same handler logic?
	// Actually we should mount them carefully. The task says /refresh for both.
	// But Fiber routes must be unique. Let's mount web auth under /web or use distinct paths.
	// Wait, we can't have two handlers for POST /refresh.
	// Let's use /web/refresh and /web/logout for now.
	webGrp := authGrp.Group("/web")
	webGrp.Post("/refresh", webAuthHandler.RefreshWeb)
	webGrp.Post("/logout", webAuthHandler.LogoutWeb)

	// Setup Middlewares
	jwtAuth := middleware.JWTAuth(signKey, encKey)

	// MFA Setup requires JWT
	authGrp.Post("/mfa/setup", jwtAuth, desktopAuthHandler.SetupMFA)

	// Admin Routes
	adminGrp := app.Group("/api/v1/admin", jwtAuth, middleware.RequireRole("admin"))

	adminUsers := adminGrp.Group("/users")
	adminUsers.Get("/", adminUserHandler.List)
	adminUsers.Post("/", adminUserHandler.Create)
	adminUsers.Get("/:id", adminUserHandler.GetByID)
	adminUsers.Put("/:id/deactivate", adminUserHandler.Deactivate)
	adminUsers.Put("/:id/roles", middleware.RequirePermission("user:assign_role", rdb, roleRepo), adminUserHandler.AssignRoles)

	adminRoles := adminGrp.Group("/roles")
	adminRoles.Get("/", adminRoleHandler.List)
	adminRoles.Get("/:id/permissions", adminRoleHandler.GetPermissions)
	adminRoles.Put("/:id/permissions", adminRoleHandler.UpdatePermissions)

	adminPerms := adminGrp.Group("/permissions")
	adminPerms.Get("/", adminRoleHandler.ListPermissions)

	adminDepts := adminGrp.Group("/departments")
	adminDepts.Get("/", adminDeptHandler.List)
	adminDepts.Post("/", adminDeptHandler.Create)

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
