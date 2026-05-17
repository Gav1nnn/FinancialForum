package global

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	// Db 是全局 GORM 连接实例。
	Db *gorm.DB
	// RedisDB 是全局 Redis 客户端实例。
	RedisDB *redis.Client
)
