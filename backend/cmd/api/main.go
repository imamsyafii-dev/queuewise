package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imamsyafii-dev/queuewise/backend/internal/cache"
	"github.com/imamsyafii-dev/queuewise/backend/internal/config"
	"github.com/imamsyafii-dev/queuewise/backend/internal/database"
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

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "queuewise-api",
			"env":      cfg.AppEnv,
			"database": "ok",
			"redis":    "ok",
		})
	})

	log.Printf("Queuewise API running on http://localhost:%s", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("failed to run API server: ", err)
	}
}
