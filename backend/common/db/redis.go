package db

import (
	"alerting-platform/common/config"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func GetRedisClient() *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	once.Do(func() {
		cfg := config.GetConfig()
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
	})

	return redisClient
}

func MockRedis(addr string) {
	if redisClient != nil {
		redisClient.Close()
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
