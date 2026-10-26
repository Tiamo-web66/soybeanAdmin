package initialize

import (
	"soybean-admin-go/global"
)

// InitAll 初始化所有组件
func InitAll() {
	// 按顺序初始化各组件

	// 1. 初始化配置
	global.CONFIG = InitConfig()

	// 2. 初始化日志
	InitLogger()

	// 3. 初始化数据库连接
	global.DB = InitDB()

	// 4. 初始化Redis客户端
	global.REDIS = InitRedis()

	// 5. 初始化gen生成的代码
	InitBO()

	// 6. 初始化RBAC权限缓存
	InitRbacCache()
}
