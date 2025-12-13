package controllers

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	r.NoRoute(authMiddleware.MiddlewareFunc(), handleNoRoute())

	v1 := r.Group("/v1")

	v1.POST("/login", authMiddleware.LoginHandler)
	v1.POST("/refresh", authMiddleware.RefreshHandler)

	authenticated := v1.Group("/", authMiddleware.MiddlewareFunc())
	{
		authenticated.POST("/logout", authMiddleware.LogoutHandler)
		authenticated.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello, authenticated user!",
			})
		})
	}
}

func handleNoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Route not found"})
	}
}
