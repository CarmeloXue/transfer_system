package redis

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	rdb  *redis.Client
	once sync.Once
)

func GetRedisClient() *redis.Client {
	once.Do(func() {
		redisHost := os.Getenv("REDIS_HOST")
		rdb = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:6379", redisHost),
		})
	})
	return rdb
}
