package middleware

import (
	"alerting-platform/api/dto"
	"alerting-platform/api/models"
	"alerting-platform/common/config"
	"log"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	jwt_go "github.com/golang-jwt/jwt/v5"
)

const identityKey = "email"

func GetJWTMiddleware() *jwt.GinJWTMiddleware {
	middleware := &jwt.GinJWTMiddleware{
		IdentityKey:     identityKey,
		Key:             []byte(config.GetConfig().Secret),
		Timeout:         time.Minute * 15,
		MaxRefresh:      time.Hour * 24,
		Authenticator:   authenticator(),
		IdentityHandler: identityHandler(),
		PayloadFunc:     payloadFunc(),
	}

	middleware.EnableRedisStore(
		jwt.WithRedisAddr(config.GetConfig().RedisHost+":"+config.GetConfig().RedisPort),
		jwt.WithRedisAuth(config.GetConfig().RedisPassword, config.GetConfig().RedisDB),
		jwt.WithRedisCache(128*1024*1024, time.Minute), // 128MB
		jwt.WithRedisKeyPrefix(config.GetConfig().RedisPrefix+":jwt:"),
	)

	middleware, err := jwt.New(middleware)
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	err = middleware.MiddlewareInit()
	if err != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + err.Error())
	}

	return middleware
}

func authenticator() func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var loginVals dto.LoginInput
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}

		email := loginVals.Email
		password := loginVals.Password

		// TODO: tu trzeba pobrać użytkownika z bazy danych. Czy to trzeba będzie cachować?
		if (email == "admin@g.com" && password == "password") || (email == "test@g.com" && password == "password") {
			return &models.User{
				Email: email,
			}, nil
		}

		return nil, jwt.ErrFailedAuthentication
	}
}

func identityHandler() func(c *gin.Context) any {
	return func(c *gin.Context) any {
		claims := jwt.ExtractClaims(c)
		return &models.User{
			Email: claims[identityKey].(string),
		}
	}
}

func payloadFunc() func(data any) jwt_go.MapClaims {
	return func(data any) jwt_go.MapClaims {
		if v, ok := data.(*models.User); ok {
			return jwt_go.MapClaims{
				identityKey: v.Email,
			}
		}
		return jwt_go.MapClaims{}
	}
}
