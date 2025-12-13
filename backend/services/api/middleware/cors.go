package middleware

import (
	"alerting-platform/common/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetCORSMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	if config.GetConfig().Env == config.DEV {
		corsConfig.AllowAllOrigins = true
		return cors.New(corsConfig)
	}

	corsConfig.AllowOrigins = []string{"https://your-production-domain.com"}

	return cors.New(corsConfig)
}
