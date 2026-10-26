package initialize

import (
	"log"
	"os"

	"soybean-admin-go/config"
)

// InitConfig 初始化配置
func InitConfig() *config.Config {
	var configPath string

	// 优先使用环境变量中的配置文件路径
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	} else {
		// 默认配置文件路径
		configPath = "config.toml"
	}

	// 加载配置文件
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
		return nil
	}

	return cfg
}
