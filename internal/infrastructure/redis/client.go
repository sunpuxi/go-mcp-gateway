package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/sunpuxi/go-mcp-gateway/config"
)

// NewRedis 根据配置创建 Redis 客户端，并验证连接可达性
func NewRedis(cfg config.RedisConfig) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 验证连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis 连接失败 %s: %w", cfg.Addr, err)
	}

	return rdb, nil
}
