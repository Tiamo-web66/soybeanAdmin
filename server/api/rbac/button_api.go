package rbac

import (
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/service/rbac"
)

type ButtonApi struct {
	ButtonService rbac.MenuService
}

func (api *ButtonApi) GetAllButtons(c *ginplus.Context) {
	buttons, err := api.ButtonService.GetAllButtons()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	rsp := make([]*vo.BtnItemVO, 0, len(buttons))
	for _, button := range buttons {
		rsp = append(rsp, &vo.BtnItemVO{
			Id:    button.ID,
			Code:  button.Code,
			Label: button.Label,
		})
	}
	c.Success(rsp)
}

func (api *ButtonApi) GetButtonIdsByRole(c *ginplus.Context) {
	var roleId int
	if c.BindRoleId(&roleId) {
		return
	}
	ids, err := api.ButtonService.GetButtonIdsByRole(roleId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(ids)
}

func (api *ButtonApi) UpdateButtonIdsByRole(c *ginplus.Context) {
	var roleId int
	var buttonIds []int
	if c.BindRoleId(&roleId) || c.BindJSON(&buttonIds) {
		return
	}
	if err := api.ButtonService.AssignRoleButtons(roleId, buttonIds); err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessMsg("更新角色按钮权限成功", nil)
}
