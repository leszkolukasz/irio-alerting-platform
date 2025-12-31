package controllers

import (
	"alerting-platform/api/db"
	"context"
	"net/http"
	"time"

	db_util "alerting-platform/common/db"
	"alerting-platform/common/db/firestore"

	"alerting-platform/api/pubsub"
	pubsub_common "alerting-platform/common/pubsub"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	PubSubService pubsub.PubSubServiceI
	Repository    db.RepositoryI
	LogRepository firestore.LogRepositoryI
}

func RegisterRoutes(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	controller := &Controller{
		PubSubService: pubsub.NewPubSubService(pubsub_common.GetClient()),
		Repository:    db.NewRepository(db.GetDBConnection()),
		LogRepository: firestore.GetLogRepository(context.Background()),
	}

	r.NoRoute(NoRouteHandler())

	r.GET("/health", HealthCheckHandler)

	v1 := r.Group("/api/v1")

	v1.POST("/login", authMiddleware.LoginHandler)
	v1.POST("/refresh", authMiddleware.RefreshHandler)
	v1.POST("/users", controller.RegisterUser)

	authenticated := v1.Group("/", authMiddleware.MiddlewareFunc())
	{
		authenticated.POST("/logout", authMiddleware.LogoutHandler)

		services := authenticated.Group("/services")
		{
			services.POST("/", controller.CreateMonitoredService)
			services.GET("/me", controller.GetMyMonitoredServices)
			services.GET("/:id", controller.GetMonitoredServiceByID)
			services.PUT("/:id", controller.UpdateMonitoredService)
			services.DELETE("/:id", controller.DeleteMonitoredService)
			services.GET("/:id/incidents", controller.GetServiceIncidents)
		}
	}
}

func NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Route not found"})
	}
}

func HealthCheckHandler(c *gin.Context) {
	conn := db.GetDBConnection()
	rawConn, err := conn.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Database connection error"})
		return
	}

	if err = rawConn.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Database ping error"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisClient := db_util.GetRedisClient()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Redis ping error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})

}
