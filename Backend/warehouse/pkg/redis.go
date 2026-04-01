package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"warehouse/internal/config"

	"github.com/redis/go-redis/v9"
)

func GetRedisConnectionPool(cfg config.RedisStorage, log *slog.Logger) *redis.Client {
    var rdb *redis.Client
    
    if cfg.URL != "" {
        // Используем URL
        opt, err := redis.ParseURL(cfg.URL)
        if err != nil {
            panic(fmt.Errorf("failed to parse Redis URL: %w", err))
        }
        rdb = redis.NewClient(opt)
    } else {
        // Fallback к старому способу
        addr := cfg.Host
        if cfg.Port != "" {
            port := strings.TrimPrefix(cfg.Port, ":")
            addr = fmt.Sprintf("%s:%s", cfg.Host, port)
        }
        
        rdb = redis.NewClient(&redis.Options{
            Addr:     addr,
            Password: cfg.Password,
            DB:       0,
        })
    }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("failed to connect to Redis: %w", err))
	}

	return rdb
}
