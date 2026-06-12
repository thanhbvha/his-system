package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thanhbvha/go-common/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"his-system/pkg/response"
)

type HealthHandler struct {
	pgPool      *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func NewHealthHandler(pgPool *pgxpool.Pool, mongoClient *mongo.Client, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		pgPool:      pgPool,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}

// CheckLiveness kiểm tra ứng dụng có đang sống không
// @Summary Liveness Check
// @Description Endpoint dùng cho K8s/Docker để biết container còn sống hay không
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) CheckLiveness(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"status":  "ok",
		"version": "1.0.0",
	})
}

// CheckReadiness kiểm tra kết nối tới tất cả dependencies (DB, Cache)
// @Summary Readiness Check
// @Description Endpoint dùng cho K8s/Docker để biết ứng dụng đã sẵn sàng nhận traffic chưa
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func (h *HealthHandler) CheckReadiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	checks := fiber.Map{
		"postgres": "ok",
		"mongodb":  "ok",
		"redis":    "ok",
	}
	isReady := true

	// Ping Postgres
	if err := h.pgPool.Ping(ctx); err != nil {
		checks["postgres"] = "fail: " + err.Error()
		isReady = false
	}

	// Ping MongoDB
	if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		checks["mongodb"] = "fail: " + err.Error()
		isReady = false
	}

	// Ping Redis
	if err := h.redisClient.Ping(ctx); err != nil {
		checks["redis"] = "fail: " + err.Error()
		isReady = false
	}

	result := fiber.Map{
		"status": "ok",
		"checks": checks,
	}

	if !isReady {
		result["status"] = "fail"
		return c.Status(fiber.StatusServiceUnavailable).JSON(result) // Không gói trong response.Fail vì chuẩn health check
	}

	return response.OK(c, result)
}
