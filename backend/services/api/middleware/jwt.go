package middleware

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"alerting-platform/api/utils"
	"alerting-platform/common/config"
	"log"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	jwt_go "github.com/golang-jwt/jwt/v5"
)

const IdentityKey = "userID"

type JWTUser struct {
	ID    uint
	Email string
}

func GetJWTMiddleware() *jwt.GinJWTMiddleware {
	middleware := &jwt.GinJWTMiddleware{
		IdentityKey:     IdentityKey,
		Key:             []byte(config.GetConfig().Secret),
		Timeout:         time.Hour * 24, // TODO; Refresh token returns empty claims for some reason.
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
		var loginVals dto.LoginRequest
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}

		email := loginVals.Email
		password := loginVals.Password

		conn := db.GetDBConnection()
		var user db.User
		result := conn.Where("email = ?", email).First(&user)
		if result.Error != nil {
			return nil, jwt.ErrFailedAuthentication
		}

		if !utils.IsPasswordHashCorrect(password, user.PasswordHash) {
			return nil, jwt.ErrFailedAuthentication
		}

		return &JWTUser{
			ID:    user.ID,
			Email: user.Email,
		}, nil
	}
}

func identityHandler() func(c *gin.Context) any {
	return func(c *gin.Context) any {
		claims := jwt.ExtractClaims(c)
		return &JWTUser{
			ID:    uint(claims[IdentityKey].(float64)),
			Email: claims["email"].(string),
		}
	}
}

func payloadFunc() func(data any) jwt_go.MapClaims {
	return func(data any) jwt_go.MapClaims {
		if v, ok := data.(*JWTUser); ok {
			return jwt_go.MapClaims{
				IdentityKey: v.ID,
				"email":     v.Email,
			}
		}

		if v, ok := data.(jwt_go.MapClaims); ok {
			return jwt_go.MapClaims{
				IdentityKey: v[IdentityKey],
				"email":     v["email"],
			}
		}

		if v, ok := data.(map[string]interface{}); ok {
			return jwt_go.MapClaims{
				IdentityKey: v[IdentityKey],
				"email":     v["email"],
			}
		}

		return jwt_go.MapClaims{}
	}
}
