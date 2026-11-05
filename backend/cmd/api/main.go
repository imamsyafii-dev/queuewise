package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imamsyafii-dev/queuewise/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration: ", err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "queuewise-api",
			"env":     cfg.AppEnv,
		})
	})

	log.Printf("Queuewise API running on http://localhost:%s", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("failed to run API server: ", err)
	}
}
