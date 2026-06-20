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

	"github.com/joho/godotenv"
	"github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/queue"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/cmd/worker/handlers"
	visitWorker "his-system/internal/visit/worker"
	"his-system/pkg/notify"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Logger
	logConfig := logger.DefaultOptions()
	logConfig.Level = -4
	appLogger := logger.New(logConfig)
	logger.SetDefault(appLogger)
	appLogger.InfoAsync("Worker Service starting...", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// Redis
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
	redisPrefix := os.Getenv("REDIS_PREFIX")
	if redisPrefix == "" {
		redisPrefix = "his_system:"
	}
	rdb := commonRedis.MustConnect(ctx, commonRedis.Config{
		Host:     host,
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
		Prefix:   redisPrefix,
	})
	defer rdb.Close()
	commonRedis.SetDefault(rdb)

	// Queue
	cfg := queue.DefaultConfig()
	cfg.Logger = appLogger
	q := queue.New(rdb, cfg)

	// --- Register existing workers ---
	zaloClient := notify.NewZaloClient()
	smsClient := notify.NewSMSClient()
	sendOTPHandler := handlers.NewSendOTPHandler(zaloClient, smsClient)
	q.RegisterJobType("send_otp", queue.JobTypeOptions{Concurrency: 5, MaxRetry: 3})
	q.RegisterHandler("send_otp", sendOTPHandler.Handle)

	// --- Register Visit domain workers ---
	visitWorker.RegisterVisitWorkers(q)

	// Start Queue
	appLogger.InfoAsync("All workers registered, starting queue...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	q.Start(ctx)
	appLogger.InfoAsync("Worker Service running", "dispatch_time", time.Now().Format(time.RFC3339Nano))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.InfoAsync("Shutting down Worker...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	cancel()
	q.Stop()
	appLogger.InfoAsync("Worker exited properly", "dispatch_time", time.Now().Format(time.RFC3339Nano))
}
