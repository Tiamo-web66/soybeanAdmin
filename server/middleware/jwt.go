package middleware

import (
	"errors"
	"soybean-admin-go/pkg/ginplus"
	"strings"

	"soybean-admin-go/utils/jwt"
)

// JWTAuth JWT验证中间件
func JWTAuth(c *ginplus.Context) bool {
	// 获取请求头中的token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.FailAuth("请求未携带token，无权限访问")
		return false
	}

	// 从Bearer中提取token
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		c.FailAuth("token格式错误")
		return false
	}

	// 解析token
	claims, err := jwt.ParseToken(parts[1])
	if err != nil {
		if errors.Is(err, jwt.TokenExpired) {
			c.FailAuth("token已过期")
			return false
		}
		c.FailAuth("token解析错误")
		return false
	}

	// 验证token类型
	if claims.Subject != string(jwt.AccessToken) {
		c.FailAuth("无效的token类型")
		return false
	}

	// 将用户信息存入上下文
	//c.Set("userName", claims.UserName)
	//c.Set("roles", claims.Roles) //这歌没有用,因为会变化,实时性不够

	c.SetUserId(int(claims.UserID))
	return true
}
