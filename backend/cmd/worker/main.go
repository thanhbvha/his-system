package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/thanhbvha/go-common/logger"
	"github.com/thanhbvha/go-common/queue"
	commonRedis "github.com/thanhbvha/go-common/redis"

	"his-system/cmd/worker/handlers"
	"his-system/pkg/notify"
)

func main() {
	if err := godotenv.Load(); err != nil {
		// Log warning but continue if no .env file
		fmt.Println("Warning: No .env file found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init Logger
	logConfig := logger.DefaultOptions()
	logConfig.Level = -4 // Debug
	appLogger := logger.New(logConfig)
	logger.SetDefault(appLogger)

	// Init Redis
	redisPrefix := os.Getenv("REDIS_PREFIX")
	if redisPrefix == "" {
		redisPrefix = "his_system:"
	}
	rdb := commonRedis.MustConnect(ctx, commonRedis.Config{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     6379,
		Password: os.Getenv("REDIS_PASSWORD"),
		Prefix:   redisPrefix,
	})
	defer rdb.Close()
	commonRedis.SetDefault(rdb)

	// Init Queue
	cfg := queue.DefaultConfig()
	q := queue.New(rdb, cfg)

	// Init Clients
	zaloClient := notify.NewZaloClient()
	smsClient := notify.NewSMSClient()

	// Init Handlers
	sendOTPHandler := handlers.NewSendOTPHandler(zaloClient, smsClient)

	// Register Jobs
	q.RegisterJobType("send_otp", queue.JobTypeOptions{Concurrency: 5, MaxRetry: 3})
	q.RegisterHandler("send_otp", sendOTPHandler.Handle)

	// Start Worker
	logger.InfoAsync("Starting Worker...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	go q.Start(ctx)

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.InfoAsync("Shutting down Worker...", "dispatch_time", time.Now().Format(time.RFC3339Nano))
	cancel() // Stop queue workers
}
