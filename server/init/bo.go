package initialize

import (
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
)

// InitBO 初始化gorm.io/gen生成的查询结构
func InitBO() {
	// 使用全局DB连接初始化bo.Q
	bo.SetDefault(global.DB)
}
