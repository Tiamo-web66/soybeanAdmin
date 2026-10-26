package api

import (
	"fmt"
	"net/http"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/test/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ButtonAPITestSuite 按钮API测试套件
type ButtonAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *ButtonAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")

	// 创建模拟API实例
	mockButtonApi := &testutil.MockButtonApi{}

	// 注册按钮管理路由
	systemManageRouterBase := apiV1.Group("/systemManage")

	// 不需要权限验证的路由
	systemManageRouter := systemManageRouterBase.Group("")
	{
		// 查询接口
		systemManageRouter.GET("/getAllButton", "获取所有按钮", "获取系统所有按钮信息", mockButtonApi.GetAllButtons)
	}

	// 需要权限验证的路由
	systemManageRouterWithAuth := systemManageRouterBase.Group("")
	{
		//角色 的按钮管理
		systemManageRouterWithAuth.GET("/role/:roleId/buttons", "获取角色的按钮ID列表", "获取指定角色的按钮ID列表", mockButtonApi.GetButtonIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/buttons", "给角色分配按钮权限", "给角色重新分配按钮权限", mockButtonApi.UpdateButtonIdsByRole)
	}

	// 生成测试用的访问令牌
	token, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "access")
	if err != nil {
		s.T().Fatalf("生成测试令牌失败: %v", err)
	}
	s.token = token
}

// TestGetAllButtons 测试获取所有按钮API
func (s *ButtonAPITestSuite) TestGetAllButtons() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getAllButton", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestGetButtonIdsByRole 测试获取角色的按钮ID列表API
func (s *ButtonAPITestSuite) TestGetButtonIdsByRole() {
	// 设置请求头，包含JWT令牌和权限验证
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/role/1/buttons", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestButtonAPI 运行按钮API测试套件
func TestButtonAPI(t *testing.T) {
	suite.Run(t, new(ButtonAPITestSuite))
}
