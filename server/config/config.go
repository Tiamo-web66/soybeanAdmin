package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config 包含应用程序所有配置项
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	Redis    RedisConfig    `toml:"redis"`
	JWT      JWTConfig      `toml:"jwt"`
	Log      LogConfig      `toml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `toml:"port"`
	Mode         string `toml:"mode"`
	ReadTimeout  int    `toml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
	SSLMode  string `toml:"sslmode"`
	MaxIdle  int    `toml:"max_idle"`
	MaxOpen  int    `toml:"max_open"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `toml:"secret"`
	Expiration int    `toml:"expiration"` // Token过期时间，单位小时
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `toml:"level"`
	Filename   string `toml:"filename"`
	MaxSize    int    `toml:"max_size"`    // 单位MB
	MaxBackups int    `toml:"max_backups"` // 最大备份数
	MaxAge     int    `toml:"max_age"`     // 最大保留天数
	Compress   bool   `toml:"compress"`    // 是否压缩
}

// LoadConfig 从指定路径加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
