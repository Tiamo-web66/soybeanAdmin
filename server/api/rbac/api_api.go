package rbac

import (
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/service/rbac"
)

type ApiApi struct {
	ApiService rbac.APIService
}

// GetAllApis 获取所有API
func (api *ApiApi) GetAllApis(c *ginplus.Context) {
	auth := c.Query("auth") //api 是否需要权限能访问
	query := bo.ApiModel.Order(bo.ApiModel.ID.Desc())
	if auth != "" {
		query = query.Where(bo.ApiModel.Auth.Eq(auth))
	}
	apis, err := query.Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(apis)
}

// GetApis 获取API列表,带分页查询
func (api *ApiApi) GetApis(c *ginplus.Context) {
	var req dto.GetApiListDTO
	if c.BindQueryPage(&req) {
		return
	}

	u := bo.ApiModel
	// 分页查询
	query := u.WithContext(global.Context())
	if req.Name != nil {
		query = query.Where(u.Name.Like("%" + *req.Name + "%"))
	}
	if req.Method != nil {
		query = query.Where(u.Path.Eq(*req.Method))
	}
	if req.Auth != nil {
		query = query.Where(u.Auth.Eq(*req.Auth))
	}
	// 计算总数
	total, err := query.Count()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	list, err := u.Scopes(req.Scope()).Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessWithPage(total, list)
}

// GetApiIdsByRole 获取角色拥有的API ID
func (api *ApiApi) GetApiIdsByRole(c *ginplus.Context) {
	var roleId int
	if c.BindRoleId(&roleId) {
		return
	}
	auth := c.Query("auth") //api 是否需要权限能访问

	ids, err := api.ApiService.GetRoleApis(roleId, auth)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(ids)
}

// UpdateApiIdsByRole 更新角色拥有的API ID
func (api *ApiApi) UpdateApiIdsByRole(c *ginplus.Context) {
	var roleId int
	var apiIds []int
	if c.BindRoleId(&roleId) || c.BindJSON(&apiIds) {
		return
	}
	// 操作人ID可根据实际需求获取，这里用0
	if err := api.ApiService.AssignRoleApis(roleId, apiIds); err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessMsg("更新角色API权限成功", nil)
}
