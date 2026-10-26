package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"soybean-admin-go/config"
	"soybean-admin-go/global"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/utils/jwt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Response 是API响应的通用结构
type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// InitTestConfig 初始化测试配置
func InitTestConfig() {
	if global.CONFIG == nil {
		global.CONFIG = &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-jwt-secret-for-unit-tests",
				Expiration: 24,
			},
		}
	}
}

// SetupRouter 创建一个测试用的路由引擎
func SetupRouter() *ginplus.Engine {
	// 确保测试配置已初始化
	InitTestConfig()
	r := ginplus.Default()
	return r
}

// GenerateTestToken 生成测试用的JWT令牌
func GenerateTestToken(userID uint, userName string, roles []string, tokenType jwt.TokenType) (string, error) {
	// 确保测试配置已初始化
	InitTestConfig()
	return jwt.GenerateToken(userID, userName, roles, tokenType)
}

// MockUserAPI 模拟用户API的行为
type MockUserAPI struct{}

// Login 模拟登录API
func (api *MockUserAPI) Login(c *ginplus.Context) {
	// 模拟成功登录响应
	c.Success(vo.LoginVO{
		Token:        "mock-access-token",
		RefreshToken: "mock-refresh-token",
	})
}

// RefreshToken 模拟刷新令牌API
func (api *MockUserAPI) RefreshToken(c *ginplus.Context) {
	// 模拟成功刷新令牌响应
	c.Success(vo.LoginVO{
		Token:        "mock-new-access-token",
		RefreshToken: "mock-new-refresh-token",
	})
}

// GetUserInfo 模拟获取用户信息API
func (api *MockUserAPI) GetUserInfo(c *ginplus.Context) {
	// 模拟成功获取用户信息响应
	c.Success(vo.UserInfoVO{
		UserId:   "1",
		UserName: "admin",
		Roles:    []string{"super-admin"},
		Buttons:  []string{"btn:add", "btn:edit", "btn:delete"},
	})
}

// UpdatePassword 模拟修改密码API
func (api *MockUserAPI) UpdatePassword(c *ginplus.Context) {
	// 模拟成功修改密码响应
	c.SuccessMsg("密码修改成功", nil)
}

// MockFrontAPI 模拟前端路由API的行为
type MockFrontAPI struct{}

// GetUserRoutes 模拟获取用户路由API
func (api *MockFrontAPI) GetUserRoutes(c *ginplus.Context) {
	// 模拟成功获取用户路由响应
	c.Success(vo.FrontRouteVO{
		Routes: []*vo.FrontRouteItemVO{
			{
				Name:      "dashboard",
				Path:      "/dashboard",
				Component: "basic",
				Children: []*vo.FrontRouteItemVO{
					{
						Name:      "dashboard_analysis",
						Path:      "analysis",
						Component: "self",
						Meta: vo.FrontRouteMetaVO{
							Title: "分析页",
							Icon:  "icon-dashboard",
						},
					},
				},
				Meta: vo.FrontRouteMetaVO{
					Title: "仪表盘",
					Icon:  "icon-dashboard",
				},
			},
		},
		Home: "dashboard_analysis",
	})
}

// Exist 模拟路由存在检查API
func (api *MockFrontAPI) Exist(c *ginplus.Context) {
	// 模拟成功检查路由存在响应
	c.Success(true)
}

// GetFixedRoutes 模拟获取固定前端路由API
func (api *MockFrontAPI) GetFixedRoutes(c *ginplus.Context) {
	// 模拟成功获取固定前端路由响应
	c.Success(vo.FrontRouteVO{
		Routes: []*vo.FrontRouteItemVO{
			{
				Name:      "login",
				Path:      "/login",
				Component: "self",
				Meta: vo.FrontRouteMetaVO{
					Title: "登录",
					Icon:  "icon-login",
				},
			},
		},
	})
}

// PerformRequest 执行HTTP请求并返回响应
func PerformRequest(r http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// 添加自定义头部
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// AssertSuccessResponse 断言响应是否成功
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder) Response {
	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "0000", response.Code)

	return response
}

// AssertErrorResponse 断言响应是否包含错误
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode string) Response {
	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedCode, response.Code)

	return response
}

// MockMenuApi 模拟菜单API的行为
type MockMenuApi struct{}

// GetAllPages 模拟获取所有页面API
func (api *MockMenuApi) GetAllPages(c *ginplus.Context) {
	// 模拟成功获取所有页面响应
	c.Success([]map[string]interface{}{
		{
			"id":        1,
			"name":      "dashboard",
			"path":      "/dashboard",
			"component": "layout/index",
		},
		{
			"id":        2,
			"name":      "system",
			"path":      "/system",
			"component": "layout/index",
		},
	})
}

// GetMenuTree 模拟获取菜单树API
func (api *MockMenuApi) GetMenuTree(c *ginplus.Context) {
	// 模拟成功获取菜单树响应
	c.Success([]map[string]interface{}{
		{
			"id":    1,
			"label": "仪表盘",
			"pId":   0,
			"children": []map[string]interface{}{
				{
					"id":    2,
					"label": "分析页",
					"pId":   1,
				},
			},
		},
	})
}

// GetMenuList 模拟获取菜单列表API
func (api *MockMenuApi) GetMenuList(c *ginplus.Context) {
	// 模拟成功获取菜单列表响应
	c.Success(map[string]interface{}{
		"total": 2,
		"records": []map[string]interface{}{
			{
				"id":        1,
				"parentId":  0,
				"menuName":  "仪表盘",
				"routeName": "dashboard",
				"routePath": "/dashboard",
				"component": "layout/index",
			},
			{
				"id":        2,
				"parentId":  1,
				"menuName":  "分析页",
				"routeName": "dashboard_analysis",
				"routePath": "analysis",
				"component": "dashboard/analysis/index",
			},
		},
	})
}

// AddMenu 模拟添加菜单API
func (api *MockMenuApi) AddMenu(c *ginplus.Context) {
	// 模拟成功添加菜单响应
	c.Success(map[string]interface{}{"id": 3})
}

// UpdateMenu 模拟更新菜单API
func (api *MockMenuApi) UpdateMenu(c *ginplus.Context) {
	// 模拟成功更新菜单响应
	c.Success(nil)
}

// DeleteMenu 模拟删除菜单API
func (api *MockMenuApi) DeleteMenu(c *ginplus.Context) {
	// 模拟成功删除菜单响应
	c.Success(nil)
}

// DeleteMenus 模拟批量删除菜单API
func (api *MockMenuApi) DeleteMenus(c *ginplus.Context) {
	// 模拟成功批量删除菜单响应
	c.Success(nil)
}

// GetMenuInfo 模拟获取菜单详情API
func (api *MockMenuApi) GetMenuInfo(c *ginplus.Context) {
	// 模拟成功获取菜单详情响应
	c.Success(map[string]interface{}{
		"id":        1,
		"parentId":  0,
		"menuName":  "仪表盘",
		"routeName": "dashboard",
		"routePath": "/dashboard",
		"component": "layout/index",
	})
}

// GetMenuIdsByRole 模拟获取角色的菜单ID列表API
func (api *MockMenuApi) GetMenuIdsByRole(c *ginplus.Context) {
	// 模拟成功获取角色的菜单ID列表响应
	c.Success([]int{1, 2})
}

// UpdateMenuIdByRole 模拟给角色分配菜单权限API
func (api *MockMenuApi) UpdateMenuIdByRole(c *ginplus.Context) {
	// 模拟成功给角色分配菜单权限响应
	c.Success(nil)
}

// MockButtonApi 模拟按钮API的行为
type MockButtonApi struct{}

// GetAllButtons 模拟获取所有按钮API
func (api *MockButtonApi) GetAllButtons(c *ginplus.Context) {
	// 模拟成功获取所有按钮响应
	c.Success([]map[string]interface{}{
		{
			"id":    1,
			"code":  "add",
			"label": "添加",
		},
		{
			"id":    2,
			"code":  "edit",
			"label": "编辑",
		},
		{
			"id":    3,
			"code":  "delete",
			"label": "删除",
		},
	})
}

// GetButtonIdsByRole 模拟获取角色的按钮ID列表API
func (api *MockButtonApi) GetButtonIdsByRole(c *ginplus.Context) {
	// 模拟成功获取角色的按钮ID列表响应
	c.Success([]int{1, 2})
}

// UpdateButtonIdsByRole 模拟给角色分配按钮权限API
func (api *MockButtonApi) UpdateButtonIdsByRole(c *ginplus.Context) {
	// 模拟成功给角色分配按钮权限响应
	c.Success(nil)
}

// MockApiApi 模拟API管理API的行为
type MockApiApi struct{}

// GetAllApis 模拟获取所有API信息的API
func (api *MockApiApi) GetAllApis(c *ginplus.Context) {
	// 模拟成功获取所有API信息响应
	c.Success([]map[string]interface{}{
		{
			"id":     1,
			"path":   "/api/v1/user/login",
			"method": "POST",
			"name":   "用户登录",
			"auth":   "0",
		},
		{
			"id":     2,
			"path":   "/api/v1/user/info",
			"method": "GET",
			"name":   "获取用户信息",
			"auth":   "1",
		},
	})
}

// GetApiIdsByRole 模拟获取角色的API ID列表API
func (api *MockApiApi) GetApiIdsByRole(c *ginplus.Context) {
	// 模拟成功获取角色的API ID列表响应
	c.Success([]int{1, 2})
}

// UpdateApiIdsByRole 模拟更新角色的API权限
func (api *MockApiApi) UpdateApiIdsByRole(c *ginplus.Context) {
	// 模拟成功更新角色的API权限响应
	c.Success(nil)
}

// MockRoleAPI 模拟角色API的行为
type MockRoleAPI struct{}

// GetRoleList 模拟获取角色列表API
func (api *MockRoleAPI) GetRoleList(c *ginplus.Context) {
	// 模拟成功获取角色列表响应
	c.Success(map[string]interface{}{
		"total": 2,
		"records": []map[string]interface{}{
			{
				"id":       1,
				"roleName": "超级管理员",
				"roleCode": "super-admin",
				"roleDesc": "系统超级管理员",
				"status":   "1",
			},
			{
				"id":       2,
				"roleName": "普通用户",
				"roleCode": "user",
				"roleDesc": "普通用户",
				"status":   "1",
			},
		},
	})
}

// GetAllRoles 模拟获取所有角色API
func (api *MockRoleAPI) GetAllRoles(c *ginplus.Context) {
	// 模拟成功获取所有角色响应
	c.Success([]map[string]interface{}{
		{
			"id":       1,
			"roleName": "超级管理员",
			"roleCode": "super-admin",
		},
		{
			"id":       2,
			"roleName": "普通用户",
			"roleCode": "user",
		},
	})
}

// CreateRole 模拟创建角色API
func (api *MockRoleAPI) CreateRole(c *ginplus.Context) {
	// 模拟成功创建角色响应
	c.Success(map[string]interface{}{"id": 3})
}

// UpdateRole 模拟更新角色API
func (api *MockRoleAPI) UpdateRole(c *ginplus.Context) {
	// 模拟成功更新角色响应
	c.Success(nil)
}

// DeleteRole 模拟删除角色API
func (api *MockRoleAPI) DeleteRole(c *ginplus.Context) {
	// 模拟成功删除角色响应
	c.Success(nil)
}

// DeleteRoles 模拟批量删除角色API
func (api *MockRoleAPI) DeleteRoles(c *ginplus.Context) {
	// 模拟成功批量删除角色响应
	c.Success(nil)
}
