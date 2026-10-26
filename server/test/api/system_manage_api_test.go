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

// SystemManageAPITestSuite 系统管理API测试套件
type SystemManageAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *SystemManageAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")

	// 创建模拟API实例
	mockMenuApi := &testutil.MockMenuApi{}
	mockRoleAPI := &testutil.MockRoleAPI{}
	mockButtonApi := &testutil.MockButtonApi{}
	mockApiApi := &testutil.MockApiApi{}

	// 注册系统管理路由
	systemManageRouterBase := apiV1.Group("/systemManage")

	// 不需要权限验证的路由
	systemManageRouter := systemManageRouterBase.Group("")
	{
		// 查询接口
		systemManageRouter.GET("/getAllPages", "获取所有页面", "获取系统所有页面组件信息", mockMenuApi.GetAllPages)
		systemManageRouter.GET("/getMenuTree", "获取菜单树", "获取系统菜单树形结构", mockMenuApi.GetMenuTree)
		systemManageRouter.GET("/getAllButton", "获取所有按钮", "获取系统所有按钮信息", mockButtonApi.GetAllButtons)
		systemManageRouter.GET("/getAllApi", "获取所有API", "获取系统所有API信息", mockApiApi.GetAllApis)
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

		// 角色管理接口
		systemManageRouterWithAuth.GET("/getRoleList", "获取角色列表", "获取系统角色列表，支持分页", mockRoleAPI.GetRoleList)
		systemManageRouterWithAuth.GET("/getAllRoles", "获取所有角色", "获取系统所有角色信息", mockRoleAPI.GetAllRoles)
		systemManageRouterWithAuth.POST("/addRole", "创建角色", "创建新的系统角色", mockRoleAPI.CreateRole)
		systemManageRouterWithAuth.PUT("/updateRole", "更新角色", "更新已有的系统角色", mockRoleAPI.UpdateRole)
		systemManageRouterWithAuth.DELETE("/deleteRole", "删除角色", "删除系统角色", mockRoleAPI.DeleteRole)
		systemManageRouterWithAuth.DELETE("/deleteRoles", "批量删除角色", "批量删除多个系统角色", mockRoleAPI.DeleteRoles)

		//角色 的菜单管理
		systemManageRouterWithAuth.GET("/role/:roleId/menus", "获取角色的菜单ID列表", "获取指定角色的菜单ID列表", mockMenuApi.GetMenuIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/menus", "给角色分配菜单权限", "给角色重新分配菜单权限", mockMenuApi.UpdateMenuIdByRole)

		//角色 的按钮管理
		systemManageRouterWithAuth.GET("/role/:roleId/buttons", "获取角色的按钮ID列表", "获取指定角色的按钮ID列表", mockButtonApi.GetButtonIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/buttons", "给角色分配按钮权限", "给角色重新分配按钮权限", mockButtonApi.UpdateButtonIdsByRole)

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

// TestGetAllPages 测试获取所有页面API
func (s *SystemManageAPITestSuite) TestGetAllPages() {
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
func (s *SystemManageAPITestSuite) TestGetMenuTree() {
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

// TestGetAllButtons 测试获取所有按钮API
func (s *SystemManageAPITestSuite) TestGetAllButtons() {
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

// TestGetAllApis 测试获取所有API信息的API
func (s *SystemManageAPITestSuite) TestGetAllApis() {
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

// TestGetMenuList 测试获取菜单列表API
func (s *SystemManageAPITestSuite) TestGetMenuList() {
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

// TestGetRoleList 测试获取角色列表API
func (s *SystemManageAPITestSuite) TestGetRoleList() {
	// 设置请求头，包含JWT令牌和权限验证
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/systemManage/getRoleList?page=1&pageSize=10", nil, headers)

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

// TestGetApiIdsByRole 测试获取角色的API ID列表API
func (s *SystemManageAPITestSuite) TestGetApiIdsByRole() {
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

// TestSystemManageAPI 运行系统管理API测试套件
func TestSystemManageAPI(t *testing.T) {
	suite.Run(t, new(SystemManageAPITestSuite))
}
