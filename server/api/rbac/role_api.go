package rbac

import (
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/po"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	userService "soybean-admin-go/service/rbac"
	"strconv"
)

type RoleAPI struct {
	RoleService userService.RoleService
}

// GetRoleList 获得角色列表
func (api *RoleAPI) GetRoleList(c *ginplus.Context) {
	var req dto.GetRoleListDTO
	if c.BindQueryPage(&req) {
		return
	}
	// 创建查询构建器
	roleQuery := bo.RoleModel.WithContext(global.Context())
	// 应用过滤条件
	if req.RoleName != "" {
		roleQuery = roleQuery.Where(bo.RoleModel.Name.Like("%" + req.RoleName + "%"))
	}

	if req.RoleCode != "" {
		roleQuery = roleQuery.Where(bo.RoleModel.Code.Like("%" + req.RoleCode + "%"))
	}

	if req.Status != nil {
		roleQuery = roleQuery.Where(bo.RoleModel.Status.Eq(*req.Status))
	}

	// 计算总数
	total, err := roleQuery.Count()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 查询分页数据
	roles, err := roleQuery.
		Preload(bo.RoleModel.CreatedBy).
		Preload(bo.RoleModel.UpdatedBy).
		Order(bo.RoleModel.ID.Desc()).
		Scopes(req.Scope()).
		Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 转换为VO
	var rsp []*vo.RoleItemVO
	for _, role := range roles {
		rsp = append(rsp, &vo.RoleItemVO{
			Id:         role.ID,
			CreateBy:   role.SafeCreateByNickName(), // 这里需要查询创建人信息，简化处理直接返回空
			CreateTime: role.CreateTime(),
			UpdateBy:   role.SafeUpdateByNickName(), // 这里需要查询更新人信息，简化处理直接返回空
			UpdateTime: role.UpdateTime(),
			Status:     role.Status,
			RoleName:   role.Name,
			RoleCode:   role.Code,
			RoleDesc:   role.Desc,
			HomeMenuId: role.HomeMenuId,
		})
	}
	c.SuccessWithPage(total, rsp)
}

// GetAllRoles 获取所有角色
func (api *RoleAPI) GetAllRoles(c *ginplus.Context) {
	// 查询所有启用的角色
	roles, err := bo.RoleModel.
		Where(bo.RoleModel.Status.Eq(po.RoleStatusEnable)).
		Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 转换为VO
	var rsp []*vo.RoleSimplifyItemVO
	for _, role := range roles {
		rsp = append(rsp, &vo.RoleSimplifyItemVO{
			Id:       strconv.Itoa(role.ID),
			RoleName: role.Name,
			RoleCode: role.Code,
		})
	}
	c.Success(rsp)
}

// CreateRole 创建角色
func (api *RoleAPI) CreateRole(c *ginplus.Context) {
	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	var req dto.CreateRoleDTO
	if c.BindJSON(&req) {
		return
	}

	// 处理创建角色
	roleID, err := api.RoleService.CreateRole(&req, operateId)
	if err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("创建角色成功", map[string]int{"roleId": roleID})
}

// UpdateRole 更新角色
func (api *RoleAPI) UpdateRole(c *ginplus.Context) {
	var req dto.UpdateRoleDTO
	if c.BindJSON(&req) {
		return
	}

	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	//不允许将角色改为超管角色
	if po.IsSuperByCode(req.RoleCode) {
		c.Fail("不允许将角色改为超管")
		return
	}

	// 处理更新角色
	if err := api.RoleService.UpdateRole(&req, operateId); err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("更新角色成功", nil)
}

// DeleteRole 删除角色
func (api *RoleAPI) DeleteRole(c *ginplus.Context) {
	// 获取角色ID
	roleIDStr := c.Query("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		c.FailParams("角色ID必须是整数")
		return
	}

	// 处理删除角色
	if err := api.RoleService.DeleteRole(roleID); err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("删除角色成功", nil)
}

// DeleteRoles 批量删除角色
func (api *RoleAPI) DeleteRoles(c *ginplus.Context) {
	roleIds, err := c.ParseArrayParamsInt("roleIds")
	if err != nil {
		c.FailParams(err.Error())
		return
	}
	if len(roleIds) == 0 {
		c.FailParams("menuIds不能为空")
		return
	}

	if err := api.RoleService.DeleteRoles(roleIds); err != nil {
		c.Fail(err.Error())
		return
	}

	c.SuccessMsg("批量删除角色成功", nil)
}
