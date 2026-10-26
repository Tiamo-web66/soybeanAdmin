package middleware

import (
	"errors"
	"log/slog"
	"soybean-admin-go/global"
	"soybean-admin-go/pkg/ginplus"
	redisUtil "soybean-admin-go/utils/redis"
)

// AuthCheck 权限验证中间件
func AuthCheck(c *ginplus.Context) bool {
	// 首先确保用户已通过JWT认证
	if !JWTAuth(c) {
		return false
	}

	// 获取当前请求的方法和路径
	method := c.Request.Method
	path := c.Request.URL.Path

	// 获取当前用户ID
	var userID int
	if exists := c.GetOperateId(&userID); !exists {
		return false
	}

	// 获取用户的角色编码
	ctx := global.Context()
	roleCodes, err := redisUtil.GetUserRoles(ctx, userID)
	if err != nil {
		if errors.Is(err, redisUtil.ErrCacheMiss) {
			// 缓存未命中，可以考虑从数据库加载
			c.FailAuth("用户角色信息未加载，请重新登录")
			return false
		}

		slog.Error("获取用户角色失败", "error", err)
		c.FailAuth("获取用户角色失败")
		return false
	}

	// 如果用户没有任何角色
	if len(roleCodes) == 0 {
		c.FailAuth("用户没有任何角色权限")
		return false
	}

	// 验证用户是否有权限访问此API
	hasPermission, err := redisUtil.VerifyApiAuth(ctx, method, path, roleCodes)
	if err != nil {
		slog.Error("验证API权限失败", "error", err)
		c.FailAuth("验证权限失败")
		return false
	}

	if !hasPermission {
		c.FailAuth("没有权限访问此API")
		return false
	}

	return true
}
