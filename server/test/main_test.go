package test

import (
	"soybean-admin-go/test/testutil"
	"testing"
)

// TestMain 是测试的入口点
func TestMain(m *testing.M) {
	// 在这里可以进行全局测试设置
	// 例如初始化测试数据库、设置环境变量等

	// 初始化测试配置
	testutil.InitTestConfig()

	// 运行所有测试
	m.Run()

	// 在这里可以进行全局测试清理
	// 例如清理测试数据库等
}
