package api

import (
	"fmt"
	"net/http"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/test/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserAPITestSuite 用户API测试套件
type UserAPITestSuite struct {
	suite.Suite
	router *ginplus.Engine
	token  string
}

// SetupSuite 在所有测试之前运行
func (s *UserAPITestSuite) SetupSuite() {
	// 创建路由引擎
	s.router = testutil.SetupRouter()

	// 初始化路由，使用模拟的API实现
	apiV1 := s.router.Group("/api/v1")
	mockUserAPI := &testutil.MockUserAPI{}

	// 注册路由
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/login", "登录", "用户登录接口", mockUserAPI.Login)
	authGroup.POST("/refreshToken", "刷新令牌", "刷新用户登录令牌", mockUserAPI.RefreshToken)

	// 需要JWT验证的路由
	authJWTGroup := authGroup.Group("")
	authJWTGroup.RequireAuth() // 标记需要认证，但在测试中不添加中间件
	authJWTGroup.GET("/getUserInfo", "获取用户信息", "获取当前登录用户的详细信息", mockUserAPI.GetUserInfo)
	authJWTGroup.POST("/password", "修改密码", "修改当前用户的登录密码", mockUserAPI.UpdatePassword)

	// 生成测试用的访问令牌
	token, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "access")
	if err != nil {
		s.T().Fatalf("生成测试令牌失败: %v", err)
	}
	s.token = token
}

// TestLogin 测试登录API
func (s *UserAPITestSuite) TestLogin() {
	// 准备登录请求数据
	loginReq := dto.LoginDTO{
		UserName: "admin",
		Password: "123456",
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodPost, "/api/v1/auth/login", loginReq, nil)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	loginResp, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证返回的令牌
	_, hasToken := loginResp["token"]
	assert.True(s.T(), hasToken, "响应中缺少token字段")
}

// TestRefreshToken 测试刷新令牌API
func (s *UserAPITestSuite) TestRefreshToken() {
	// 准备刷新令牌请求数据
	refreshToken, err := testutil.GenerateTestToken(1, "admin", []string{"super-admin"}, "refresh")
	if err != nil {
		s.T().Fatalf("生成刷新令牌失败: %v", err)
	}

	refreshReq := dto.RefreshTokenReq{
		RefreshToken: refreshToken,
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodPost, "/api/v1/auth/refreshToken", refreshReq, nil)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	refreshResp, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证返回的令牌
	_, hasToken := refreshResp["token"]
	assert.True(s.T(), hasToken, "响应中缺少token字段")
}

// TestGetUserInfo 测试获取用户信息API
func (s *UserAPITestSuite) TestGetUserInfo() {
	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodGet, "/api/v1/auth/getUserInfo", nil, headers)

	// 断言响应
	response := testutil.AssertSuccessResponse(s.T(), w)

	// 验证返回的数据结构
	userInfo, ok := response.Data.(map[string]interface{})
	assert.True(s.T(), ok, "响应数据类型不正确")

	// 验证用户信息字段
	_, hasUserName := userInfo["userName"]
	assert.True(s.T(), hasUserName, "响应中缺少userName字段")
}

// TestUpdatePassword 测试修改密码API
func (s *UserAPITestSuite) TestUpdatePassword() {
	// 准备修改密码请求数据
	updatePasswordReq := dto.UpdatePasswordDTO{
		OldPassword: "123456",
		NewPassword: "newpassword123",
	}

	// 设置请求头，包含JWT令牌
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.token),
	}

	// 执行请求
	w := testutil.PerformRequest(s.router, http.MethodPost, "/api/v1/auth/password", updatePasswordReq, headers)

	// 断言响应
	testutil.AssertSuccessResponse(s.T(), w)
}

// TestUserAPI 运行用户API测试套件
func TestUserAPI(t *testing.T) {
	suite.Run(t, new(UserAPITestSuite))
}
