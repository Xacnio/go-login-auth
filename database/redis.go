package database

import (
	"github.com/go-redis/redis/v7"
	"os"
)

var RedisClient *redis.Client

func ConnectRedis() {
	/* Connect to Redis Server */
	dsn := os.Getenv("REDIS_ADDR")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}
