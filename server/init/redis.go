package initialize

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"soybean-admin-go/global"
)

// InitRedis 初始化Redis客户端
func InitRedis() *redis.Client {
	cfg := global.CONFIG.Redis

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("初始化Redis客户端失败: %v", err)
		return nil
	}

	return client
}
