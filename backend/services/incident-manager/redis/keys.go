package redis

import (
	"alerting-platform/common/config"
	"strconv"
)

func GetDownSinceKey(serviceID uint64) string {
	cfg := config.GetConfig()
	return cfg.RedisPrefix + ":service:" + strconv.FormatUint(serviceID, 10) + ":down_since"
}

func GetIncidentKey(serviceID uint64) string {
	cfg := config.GetConfig()
	return cfg.RedisPrefix + ":service:" + strconv.FormatUint(serviceID, 10) + ":incident"
}

func GetServiceStatusKey(serviceID uint64) string {
	return "common:service:" + strconv.FormatUint(serviceID, 10) + ":status"
}

func GetOncallerDeadlineSetKey() string {
	cfg := config.GetConfig()
	return cfg.RedisPrefix + ":oncaller_deadlines"
}
