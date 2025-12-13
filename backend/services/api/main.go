package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"alerting-platform/api/controllers"
	"alerting-platform/api/middleware"
	"alerting-platform/common/config"
)

func main() {
	if config.GetConfig().Env == config.PROD {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(middleware.GetSecurityMiddleware())
	router.Use(middleware.GetCORSMiddleware())

	authMiddleware := middleware.GetJWTMiddleware()

	controllers.RegisterRoutes(router, authMiddleware)

	port := config.GetConfig().Port
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatal(err)
	}
}
