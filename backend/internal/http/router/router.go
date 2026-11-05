package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imamsyafii-dev/queuewise/backend/internal/config"
	"github.com/imamsyafii-dev/queuewise/backend/internal/http/handler"
	"github.com/imamsyafii-dev/queuewise/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(db *pgxpool.Pool, cfg config.Config) *gin.Engine {
	engine := gin.Default()
	engine.Use(cors())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok", "service": "queuewise-api", "env": cfg.AppEnv,
			"database": "ok", "redis": "ok",
		})
	})

	queuewiseHandler := handler.NewQueuewiseHandler(repository.NewQueuewiseRepository(db))
	public := engine.Group("/api/public")
	{
		public.GET("/catalog", queuewiseHandler.Catalog)
		public.POST("/appointments", queuewiseHandler.CreateAppointment)
		public.GET("/appointments/:bookingCode", queuewiseHandler.Appointment)
		public.GET("/queue/:branchSlug", queuewiseHandler.QueueBoard)
	}
	admin := engine.Group("/api/admin")
	{
		admin.GET("/appointments", queuewiseHandler.ListAppointments)
		admin.POST("/appointments/:appointmentID/check-in", queuewiseHandler.CheckIn)
		admin.POST("/queue/:queueID/:action", queuewiseHandler.QueueAction)
	}
	return engine
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}
