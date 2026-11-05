package main

import (
	"context"
	"log"

	"github.com/imamsyafii-dev/queuewise/backend/internal/cache"
	"github.com/imamsyafii-dev/queuewise/backend/internal/config"
	"github.com/imamsyafii-dev/queuewise/backend/internal/database"
	"github.com/imamsyafii-dev/queuewise/backend/internal/http/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration: ", err)
	}

	ctx := context.Background()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL: ", err)
	}
	defer db.Close()

	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatal("failed to connect to Redis: ", err)
	}
	defer redisClient.Close()

	router := router.New(db, cfg)

	log.Printf("Queuewise API running on http://localhost:%s", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("failed to run API server: ", err)
	}
}
