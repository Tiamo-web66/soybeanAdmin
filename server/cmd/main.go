package main

import (
	"fmt"
	"log/slog"
	"soybean-admin-go/global"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/router"
	"soybean-admin-go/service/rbac"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化所有组件
	initialize.InitAll()

	// 设置Gin运行模式
	gin.SetMode(global.CONFIG.Server.Mode)

	// 创建Gin引擎
	r := ginplus.Default()

	// 初始化路由
	router.InitRouter(r)

	// 创建API服务实例
	apiService := rbac.NewAPIService()

	// 同步路由到数据库
	err := apiService.SyncAPIToDatabase(r, 1)
	if err != nil {
		slog.Error("同步路由到数据库失败", "error", err)
		return
	}

	// 启动服务
	addr := fmt.Sprintf(":%d", global.CONFIG.Server.Port)
	slog.Info("服务器启动成功", "port", global.CONFIG.Server.Port)
	if err := r.Run(addr); err != nil {
		slog.Error("服务器启动失败", "error", err)
	}
}
