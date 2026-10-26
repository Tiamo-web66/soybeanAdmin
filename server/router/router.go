package router

import (
	"soybean-admin-go/pkg/ginplus"
)

// InitRouter 初始化所有路由
func InitRouter(r *ginplus.Engine) {
	// API版本路由组
	apiV1 := r.Group("/api/v1")
	// 用户路由组
	InitUserRouter(apiV1)
	// 前端路由组
	InitFrontRouter(apiV1)
	// 系统管理路由组
	InitSystemManageRouter(apiV1)
}
