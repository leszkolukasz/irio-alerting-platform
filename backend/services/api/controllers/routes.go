package controllers

import (
	"alerting-platform/api/db"
	"context"
	"net/http"
	"time"

	db_util "alerting-platform/common/db"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	r.NoRoute(NoRouteHandler())

	r.GET("/health", HealthCheckHandler)

	v1 := r.Group("/api/v1")

	v1.POST("/login", authMiddleware.LoginHandler)
	v1.POST("/refresh", authMiddleware.RefreshHandler)
	v1.POST("/users", RegisterUser)

	authenticated := v1.Group("/", authMiddleware.MiddlewareFunc())
	{
		authenticated.POST("/logout", authMiddleware.LogoutHandler)

		services := authenticated.Group("/services")
		{
			services.POST("/", CreateMonitoredService)
			services.GET("/me", GetMyMonitoredServices)
			services.GET("/:id", GetMonitoredServiceByID)
			services.PUT("/:id", UpdateMonitoredService)
			services.DELETE("/:id", DeleteMonitoredService)
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
