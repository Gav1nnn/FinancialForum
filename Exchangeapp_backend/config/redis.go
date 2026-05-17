package config

import (
	"exchangeapp/global"
	"log"

	"github.com/go-redis/redis"
)

// InitRedis 初始化 Redis 客户端并写入全局变量。
func InitRedis() {

	addr := AppConfig.Redis.Addr
	db := AppConfig.Redis.DB
	password := AppConfig.Redis.Password

	RedisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis, got error: %v", err)
	}

	global.RedisDB = RedisClient
}
