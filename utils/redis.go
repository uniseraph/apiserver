package utils

import "fmt"

func RedisKey(k string, prefix string) (key string) {
	key = fmt.Sprintf("%s.%s.%s", KEY_REDIS_UID, prefix, k)
	return
}

func RedisSessionKey(k string) (key string) {
	return RedisKey(k, "session")
}

func ContainerAuditSessionKey(k string) (key string) {
	return RedisKey(k, "ca")
}
