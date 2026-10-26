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

// FrontAPITestSuite 前端路由API测试套件
type FrontAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *FrontAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")
	mockFrontAPI := &testutil.MockFrontAPI{}

	// 注册路由
	routeGroup := apiV1.Group("/route")
	routeGroup.GET("/isRouteExist", "前端路由存在检查", "检查路由是否存在", mockFrontAPI.Exist)
	routeGroup.GET("/getConstantRoutes", "获取固定前端路由", "获取系统固定的路由数据", mockFrontAPI.GetFixedRoutes)

	// 需要JWT验证的路由
	routeAuthGroup := routeGroup.Group("")
	routeAuthGroup.RequireAuth() // 标记需要认证，但在测试中不添加中间件
	routeAuthGroup.GET("/getUserRoutes", "获取用户前端路由", "获取当前用户可访问的路由数据", mockFrontAPI.GetUserRoutes)

	// 生成测试用的访问令牌
	token, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "access")
	if err != nil {
		s.T().Fatalf("生成测试令牌失败: %v", err)
	}
	s.token = token
}

// TestExist 测试路由存在检查API
func (s *FrontAPITestSuite) TestExist() {
	// 构造查询参数
	path := "/api/v1/route/isRouteExist?name=dashboard"

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, path, nil, nil)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	exists, ok := response.Data.(bool)
	assert.True(s.T(), ok, "响应数据类型不正确")
	assert.True(s.T(), exists, "路由应该存在")
}

// TestGetFixedRoutes 测试获取固定前端路由API
func (s *FrontAPITestSuite) TestGetFixedRoutes() {
	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/route/getConstantRoutes", nil, nil)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	routes, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证routes字段
	_, hasRoutes := routes["routes"]
	assert.True(s.T(), hasRoutes, "响应中缺少routes字段")
}

// TestGetUserRoutes 测试获取用户前端路由API
func (s *FrontAPITestSuite) TestGetUserRoutes() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/route/getUserRoutes", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	routes, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证routes字段
	_, hasRoutes := routes["routes"]
	assert.True(s.T(), hasRoutes, "响应中缺少routes字段")
}

// TestFrontAPI 运行前端路由API测试套件
func TestFrontAPI(t *testing.T) {
	suite.Run(t, new(FrontAPITestSuite))
}
