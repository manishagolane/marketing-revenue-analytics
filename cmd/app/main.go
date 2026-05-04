package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marketing-revenue-analytics/auth"
	"marketing-revenue-analytics/cache"
	"marketing-revenue-analytics/config"
	database "marketing-revenue-analytics/db"
	"marketing-revenue-analytics/internal/clients"
	"marketing-revenue-analytics/internal/consumer"
	"marketing-revenue-analytics/internal/server"
	"marketing-revenue-analytics/logger"
	"marketing-revenue-analytics/models"

	"go.uber.org/zap"
)

func main() {
	// 1. Config
	config.Init()

	// 2. Logger
	appLogger := logger.Init()
	defer appLogger.Sync()

	// 3. Database
	pool := database.Init(appLogger)
	queries := models.New(pool)

	appClients := clients.InitializeClients(appLogger, queries, pool)

	// 4. Cache
	cacheClient := cache.NewCache(queries, appLogger)

	// 5. JWT
	jwtManager := auth.NewJWTManager(appLogger)

	// 6. Server config from viper
	cfg := server.Config{
		Port:         config.GetString("server.port"),
		Environment:  config.GetString("environment"),
		ReadTimeout:  config.GetInt("server.readTimeout"),
		WriteTimeout: config.GetInt("server.writeTimeout"),
	}

	app, err := server.New(cfg, pool, queries, jwtManager, cacheClient, appLogger, appClients)
	if err != nil {
		appLogger.Fatal("failed to build server", zap.Error(err))
	}

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	eventConsumer := consumer.NewEventConsumer(queries, pool, appLogger)
	queueURL := config.GetString("aws.sqs")
	go appClients.SqsClient.StartConsuming(consumerCtx, queueURL, eventConsumer.Consume)

	errCh := make(chan error, 1)
	go func() { errCh <- app.Start() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		appLogger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			appLogger.Fatal("server error", zap.Error(err))
		}
	}

	cancelConsumer() // stops SQS polling loop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		appLogger.Error("shutdown error", zap.Error(err))
		os.Exit(1)
	}
}
