package redis

import "strconv"

func GetServiceStatusKey(serviceID uint64) string {
	return "common:service:" + strconv.FormatUint(serviceID, 10) + ":status"
}
