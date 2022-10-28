package rdcon

import (
	"github.com/go-redis/redis/v7"
	"sync"
	"time"
)

type Redis struct {
	*redis.Client
}

var (
	redisInstance *Redis
	once          sync.Once
)

func GetRedis() *Redis {
	once.Do(func() {
		//cfg := config.Redis

		redisInstance = &Redis{
			Client: redis.NewClient(&redis.Options{
				Addr:         "localhost:6379",
				DialTimeout:  10 * time.Second,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				PoolSize:     2,
				PoolTimeout:  30 * time.Second,
			}),
		}
	})
	return redisInstance
}
