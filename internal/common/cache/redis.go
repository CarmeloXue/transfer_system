package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	ctx := context.Background()

	// Initialize a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redis service name in Docker Compose
		Password: "",           // No password set
		DB:       0,            // Use default DB
	})

	// Test the connection by pinging the Redis server
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return nil, err
	}

	return rdb, nil
}
