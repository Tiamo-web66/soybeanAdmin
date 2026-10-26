package global

import (
	"context"
	"soybean-admin-go/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	// DB 全局数据库连接
	DB *gorm.DB

	// REDIS 全局Redis客户端
	REDIS *redis.Client

	// CONFIG 全局配置
	CONFIG *config.Config
)

// Context 返回一个上下文，用于数据库操作
func Context() context.Context {
	return context.Background()
}
