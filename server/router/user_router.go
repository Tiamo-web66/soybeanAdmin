package router

import (
	"soybean-admin-go/api/rbac"
	"soybean-admin-go/middleware"
	"soybean-admin-go/pkg/ginplus"
)

// InitUserRouter 初始化菜单路由
func InitUserRouter(r *ginplus.RouterGroup) {
	// 用户路由
	userRouter := r.Group("/auth")
	userAPI := rbac.UserAPI{}
	otherAPI := rbac.OtherAPI{}

	// 公开路由
	userRouter.POST("/login", "登录", "用户登录接口", userAPI.Login)
	userRouter.POST("/refreshToken", "刷新令牌", "刷新用户登录令牌", userAPI.RefreshToken)
	userRouter.GET("/error", "自定义错误", "前端传递自定义错误,后端整理好后返回", otherAPI.CustomError)

	// 需要JWT验证的路由
	userRouterAuth := userRouter.Group("")
	userRouterAuth.Use(middleware.JWTAuth).RequireAuth()
	{
		userRouterAuth.GET("/getUserInfo", "获取用户信息", "获取当前登录用户的详细信息", userAPI.GetUserInfo)
		userRouterAuth.POST("/password", "修改密码", "修改当前用户的登录密码", userAPI.UpdatePassword)
	}

	// 需要权限验证的路由
	userRouterWithAuth := userRouter.Group("")
	userRouterWithAuth.Use(middleware.AuthCheck).RequireAuth()
	{
		// 这些API需要权限验证
		// 可以根据需要添加需要权限验证的API
	}
}
