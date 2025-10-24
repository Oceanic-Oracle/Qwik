package pkg

import (
	"auth/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func GetRedisConnectionPool(cfg config.RedisStorage, log *slog.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Host + cfg.Port,
		Password:     cfg.Password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		MaxRetries:   3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("failed to connect to Redis: %w", err))
	}

	return rdb
}
