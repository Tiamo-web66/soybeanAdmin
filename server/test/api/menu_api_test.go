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

// MenuAPITestSuite 菜单API测试套件
type MenuAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *MenuAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")

	// 创建模拟API实例
	mockMenuApi := &testutil.MockMenuApi{}

	// 注册菜单管理路由
	systemManageRouterBase := apiV1.Group("/systemManage")

	// 不需要权限验证的路由
	systemManageRouter := systemManageRouterBase.Group("")
	{
		// 查询接口
		systemManageRouter.GET("/getAllPages", "获取所有页面", "获取系统所有页面组件信息", mockMenuApi.GetAllPages)
		systemManageRouter.GET("/getMenuTree", "获取菜单树", "获取系统菜单树形结构", mockMenuApi.GetMenuTree)
	}

	// 需要权限验证的路由
	systemManageRouterWithAuth := systemManageRouterBase.Group("")
	{
		// 菜单管理接口
		systemManageRouterWithAuth.GET("/getMenuList/v2", "获取菜单列表", "获取系统菜单列表，支持分页", mockMenuApi.GetMenuList)
		systemManageRouterWithAuth.POST("/addMenu", "添加菜单", "添加新的系统菜单", mockMenuApi.AddMenu)
		systemManageRouterWithAuth.PUT("/updateMenu", "更新菜单", "更新已有的系统菜单", mockMenuApi.UpdateMenu)
		systemManageRouterWithAuth.DELETE("/deleteMenu", "删除菜单", "删除单个系统菜单", mockMenuApi.DeleteMenu)
		systemManageRouterWithAuth.DELETE("/deleteMenus", "批量删除菜单", "批量删除多个系统菜单", mockMenuApi.DeleteMenus)
		systemManageRouterWithAuth.GET("/menuInfo/:id", "获取菜单详情", "获取指定ID的菜单详细信息", mockMenuApi.GetMenuInfo)

		//角色 的菜单管理
		systemManageRouterWithAuth.GET("/role/:roleId/menus", "获取角色的菜单ID列表", "获取指定角色的菜单ID列表", mockMenuApi.GetMenuIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/menus", "给角色分配菜单权限", "给角色重新分配菜单权限", mockMenuApi.UpdateMenuIdByRole)
	}

	// 生成测试用的访问令牌
	token, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "access")
	if err != nil {
		s.T().Fatalf("生成测试令牌失败: %v", err)
	}
	s.token = token
}

// TestGetAllPages 测试获取所有页面API
func (s *MenuAPITestSuite) TestGetAllPages() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getAllPages", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestGetMenuTree 测试获取菜单树API
func (s *MenuAPITestSuite) TestGetMenuTree() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getMenuTree", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestGetMenuList 测试获取菜单列表API
func (s *MenuAPITestSuite) TestGetMenuList() {
	// 设置请求头，包含JWT令牌和权限验证
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getMenuList/v2?page=1&pageSize=10", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	data, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证分页字段
	_, hasTotal := data["total"]
	assert.True(s.T(), hasTotal, "响应中缺少total字段")

	_, hasRecords := data["records"]
	assert.True(s.T(), hasRecords, "响应中缺少records字段")
}

// TestGetMenuIdsByRole 测试获取角色的菜单ID列表API
func (s *MenuAPITestSuite) TestGetMenuIdsByRole() {
	// 设置请求头，包含JWT令牌和权限验证
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/role/1/menus", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据是否为数组
	_, ok := response.Data.([]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确，应为数组")
}

// TestMenuAPI 运行菜单API测试套件
func TestMenuAPI(t *testing.T) {
	suite.Run(t, new(MenuAPITestSuite))
}
