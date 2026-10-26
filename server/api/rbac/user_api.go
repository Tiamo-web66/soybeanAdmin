package rbac

import (
	"github.com/gin-gonic/gin"
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	userService "soybean-admin-go/service/rbac"

	"strconv"
)

// UserAPI 用户API控制器
type UserAPI struct {
	userService userService.UserService
}

// Login 用户登录
func (api *UserAPI) Login(c *ginplus.Context) {
	var req dto.LoginDTO
	if c.BindJSON(&req) {
		return
	}
	// 处理登录
	resp, err := api.userService.Login(&req)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(resp)
}

// RefreshToken 刷新令牌
func (api *UserAPI) RefreshToken(c *ginplus.Context) {
	var req dto.RefreshTokenReq
	if c.BindJSON(&req) {
		return
	}

	// 处理刷新令牌
	resp, err := api.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(resp)
}

// GetUserInfo 获取用户信息
func (api *UserAPI) GetUserInfo(c *ginplus.Context) {
	var userId int
	if exists := c.GetOperateId(&userId); !exists {
		return
	}

	// 获取用户信息
	resp, err := api.userService.GetUserInfo(userId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(resp)
}

// UpdatePassword 修改密码
func (api *UserAPI) UpdatePassword(c *ginplus.Context) {
	var userID int
	if exists := c.GetOperateId(&userID); !exists {
		return
	}
	var req dto.UpdatePasswordDTO
	if c.BindJSON(&req) {
		return
	}

	// 处理密码修改
	err := api.userService.UpdateUserPassword(userID, &req)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessMsg("密码修改成功", nil)
}

// GetUserList 获取用户列表
func (api *UserAPI) GetUserList(c *ginplus.Context) {
	var req dto.GetUserListDTO
	if c.BindQueryPage(&req) {
		return
	}

	// 创建查询构建器
	userQuery := bo.UserModel.WithContext(global.Context())

	// 应用过滤条件
	if req.UserName != "" {
		userQuery = userQuery.Where(bo.UserModel.UserName.Like("%" + req.UserName + "%"))
	}

	if req.UserGender != nil {
		userQuery = userQuery.Where(bo.UserModel.UserGender.Eq(*req.UserGender))
	}

	if req.NickName != "" {
		userQuery = userQuery.Where(bo.UserModel.NickName.Like("%" + req.NickName + "%"))
	}

	if req.UserPhone != "" {
		userQuery = userQuery.Where(bo.UserModel.UserPhone.Like("%" + req.UserPhone + "%"))
	}

	if req.UserEmail != "" {
		userQuery = userQuery.Where(bo.UserModel.UserEmail.Like("%" + req.UserEmail + "%"))
	}

	if req.UserStatus != nil {
		userQuery = userQuery.Where(bo.UserModel.Status.Eq(*req.UserStatus))
	}

	// 计算总数
	total, err := userQuery.Count()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 查询分页数据
	users, err := userQuery.
		Preload(bo.UserModel.Roles).
		Preload(bo.UserModel.CreatedBy).
		Preload(bo.UserModel.UpdatedBy).
		Order(bo.UserModel.ID.Desc()).
		Scopes(req.Scope()).
		Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 转换为VO
	var rsp []*vo.UserItemVO
	for _, user := range users {
		// 提取角色编码
		roles := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roles = append(roles, role.Code)
		}

		userVO := &vo.UserItemVO{
			Id:         user.ID,
			CreateBy:   user.SafeCreateByNickName(),
			CreateTime: user.CreateTime(),
			UpdateBy:   user.SafeUpdateByNickName(),
			UpdateTime: user.UpdateTime(),
			Status:     user.Status, // Status 已经是 string 类型
			UserName:   user.UserName,
			UserGender: user.UserGender, // UserGender 已经是 string 类型
			NickName:   user.NickName,
			UserPhone:  user.UserPhone,
			UserEmail:  user.UserEmail,
			UserRoles:  roles,
		}
		rsp = append(rsp, userVO)
	}
	c.SuccessWithPage(total, rsp)
}

// CreateUser 创建用户
func (api *UserAPI) CreateUser(c *ginplus.Context) {
	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	var req dto.CreateUserDTO
	if c.BindJSON(&req) {
		return
	}

	// 创建用户，默认密码123456
	userID, err := api.userService.CreateUser(&req, operateId)
	if err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("用户创建成功", gin.H{"id": userID})
}

// UpdateUser 修改用户
func (api *UserAPI) UpdateUser(c *ginplus.Context) {
	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	var req dto.UpdateUserDTO
	if c.BindJSON(&req) {
		return
	}

	// 处理用户更新
	err := api.userService.UpdateUser(&req, operateId)
	if err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("用户信息更新成功", nil)
}

// DeleteUser 删除用户
func (api *UserAPI) DeleteUser(c *ginplus.Context) {
	userIdStr := c.Query("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.FailParams("用户ID必须是整数")
		return
	}

	// 执行删除操作
	err = api.userService.DeleteUser(userId)
	if err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("用户删除成功", nil)
}

// DeleteUsers 批量删除用户
func (api *UserAPI) DeleteUsers(c *ginplus.Context) {
	userIds, err := c.ParseArrayParamsInt("userIds")
	if err != nil {
		c.FailParams(err.Error())
		return
	}

	if len(userIds) == 0 {
		c.FailParams("用户ID列表不能为空")
		return
	}

	// 执行批量删除
	err = api.userService.DeleteUsers(userIds)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessMsg("批量删除用户成功", nil)
}
