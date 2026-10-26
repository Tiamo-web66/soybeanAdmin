package initialize

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"soybean-admin-go/global"

	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化日志系统，使用log/slog并将日志写入文件
func InitLogger() {
	cfg := global.CONFIG.Log

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 按天切割日志的文件名
	dailyLogFile := &lumberjack.Logger{
		Filename:   getLogFileName(cfg.Filename),
		MaxSize:    cfg.MaxSize, // 以MB为单位
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge, // 以天为单位
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	// 创建多输出写入器，同时写入文件和标准输出
	multi := io.MultiWriter(dailyLogFile, os.Stdout)

	// 创建slog处理器
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     getLevelFromString(cfg.Level),
		AddSource: true, // 添加调用点信息
	}

	// 使用JSON格式的日志
	handler = slog.NewJSONHandler(multi, opts)

	// 创建slog日志记录器并设置为全局日志记录器
	logger := slog.New(handler)

	// 设置为默认的slog
	slog.SetDefault(logger)

	// 为了兼容性，也将标准库log指向我们的日志文件
	log.SetOutput(multi)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("[BIU-GO] ")

	// 使用新的slog记录日志
	slog.Info("日志系统初始化完成", "logger", "slog")
}

// getLogFileName 生成基于日期的日志文件名
func getLogFileName(baseFilename string) string {
	today := time.Now().Format("2006-01-02")
	ext := filepath.Ext(baseFilename)
	name := baseFilename[:len(baseFilename)-len(ext)]
	return fmt.Sprintf("%s.%s%s", name, today, ext)
}

// getLevelFromString 将配置中的日志级别字符串转换为slog.Level
func getLevelFromString(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
