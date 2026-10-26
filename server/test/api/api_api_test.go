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

// ApiAPITestSuite API管理API测试套件
type ApiAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *ApiAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")

	// 创建模拟API实例
	mockApiApi := &testutil.MockApiApi{}

	// 注册API管理路由
	systemManageRouterBase := apiV1.Group("/systemManage")

	// 不需要权限验证的路由
	systemManageRouter := systemManageRouterBase.Group("")
	{
		// 查询接口
		systemManageRouter.GET("/getAllApi", "获取所有API", "获取系统所有API信息", mockApiApi.GetAllApis)
	}

	// 需要权限验证的路由
	systemManageRouterWithAuth := systemManageRouterBase.Group("")
	{
		//角色 的API管理
		systemManageRouterWithAuth.GET("/role/:roleId/apis", "获取角色的API ID列表", "获取指定角色的API ID列表", mockApiApi.GetApiIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/apis", "给角色分配API权限", "给角色重新分配API权限", mockApiApi.UpdateApiIdsByRole)
	}

	// 生成测试用的访问令牌
	token, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "access")
	if err != nil {
		s.T().Fatalf("生成测试令牌失败: %v", err)
	}
	s.token = token
}

// TestGetAllApis 测试获取所有API信息的API
func (s *ApiAPITestSuite) TestGetAllApis() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getAllApi", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestGetApiIdsByRole 测试获取角色的API ID列表API
func (s *ApiAPITestSuite) TestGetApiIdsByRole() {
	// 设置请求头，包含JWT令牌和权限验证
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/role/1/apis", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestApiAPI 运行API管理API测试套件
func TestApiAPI(t *testing.T) {
	suite.Run(t, new(ApiAPITestSuite))
}
