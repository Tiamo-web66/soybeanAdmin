package router

import (
	"soybean-admin-go/api/rbac"
	"soybean-admin-go/middleware"
	"soybean-admin-go/pkg/ginplus"
)

// InitFrontRouter 初始化前端路由
func InitFrontRouter(r *ginplus.RouterGroup) {
	frontRouter := r.Group("/route")

	frontAPI := rbac.FrontAPI{}

	// 公开路由
	frontRouter.GET("/isRouteExist", "前端路由存在检查", "检查路由是否存在", frontAPI.Exist)
	frontRouter.GET("/getConstantRoutes", "获取固定前端路由", "获取系统固定的路由数据", frontAPI.GetFixedRoutes)

	// 需要JWT验证的路由
	userRouterAuth := frontRouter.Group("")
	userRouterAuth.Use(middleware.JWTAuth).RequireAuth()
	{
		userRouterAuth.GET("/getUserRoutes", "获取用户前端路由", "获取当前用户可访问的路由数据", frontAPI.GetUserRoutes)
	}
}
