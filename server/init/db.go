package initialize

import (
	"context"
	"fmt"
	"log/slog"
	"soybean-admin-go/model/po"
	"time"

	"soybean-admin-go/global"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
func InitDB() *gorm.DB {
	cfg := global.CONFIG.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// 配置GORM日志
	gormLogger := logger.New(
		NewGormLogger(), // 自定义日志接口实现
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢SQL阈值
			LogLevel:                  logger.Info,            // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound错误
			Colorful:                  false,                  // 禁用彩色
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		slog.Error("初始化数据库连接失败", "error", err)
		return nil
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("获取数据库连接池失败", "error", err)
		return nil
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)

	// 自动迁移数据库表结构
	err = db.AutoMigrate(
		&po.UserModel{},
		&po.RoleModel{},
		&po.MenuModel{},
		&po.ApiModel{},
		&po.ButtonModel{},
	)
	if err != nil {
		slog.Error("数据库自动迁移失败", "error", err)
		return nil
	}

	slog.Info("数据库连接初始化成功")
	return db
}

// NewGormLogger 创建一个新的GormLogger实例
func NewGormLogger() logger.Writer {
	return &GormWriter{}
}

// GormWriter 实现logger.Writer接口
type GormWriter struct{}

// Printf 实现logger.Writer接口
func (l *GormWriter) Printf(format string, args ...interface{}) {
	// 使用slog记录SQL日志
	msg := fmt.Sprintf(format, args...)
	slog.Info(msg, "source", "gorm")
}

// GormTrace Trace 自定义SQL日志追踪函数
func GormTrace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 创建结构化日志属性
	attrs := []any{
		slog.Duration("elapsed", elapsed),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		slog.Error("SQL执行失败", attrs...)
		return
	}

	// 记录慢SQL
	if elapsed > 200*time.Millisecond {
		slog.Warn("慢SQL执行", attrs...)
		return
	}

	// 记录正常SQL
	slog.Debug("SQL执行", attrs...)
}
