package shared

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

func NewRedis() *redis.Client {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   0,
	})
}
