package rbac

import (
	"github.com/gin-gonic/gin"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/service/rbac"
	"strconv"
)

type MenuApi struct {
	menuService rbac.MenuService
}

// AddMenu 添加菜单
func (api *MenuApi) AddMenu(c *ginplus.Context) {
	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	var req dto.AddMenuDTO
	if c.BindJSON(&req) {
		return
	}
	// 调用服务添加菜单
	id, err := api.menuService.AddMenu(&req, operateId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(gin.H{"id": id})
}

// UpdateMenu 修改菜单
func (api *MenuApi) UpdateMenu(c *ginplus.Context) {
	var operateId int
	if exists := c.GetOperateId(&operateId); !exists {
		return
	}
	var req dto.EditMenuDTO
	if c.BindJSON(&req) {
		return
	}

	// 调用服务更新菜单
	err := api.menuService.UpdateMenu(&req, operateId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(nil)
}

// DeleteMenu 删除菜单
func (api *MenuApi) DeleteMenu(c *ginplus.Context) {
	menuIdStr := c.Query("menuId")
	if menuIdStr == "" {
		c.FailParams("menuId不能为空")
		return
	}
	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		c.FailParams("menuId应该是int类型")
		return
	}

	// 调用服务删除菜单
	err = api.menuService.DeleteMenu(menuId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(nil)
}

// DeleteMenus 批量删除菜单
func (api *MenuApi) DeleteMenus(c *ginplus.Context) {
	menuIds, err := c.ParseArrayParamsInt("menuIds")
	if err != nil {
		c.FailParams(err.Error())
		return
	}
	if len(menuIds) == 0 {
		c.FailParams("menuIds不能为空")
		return
	}

	// 调用服务批量删除菜单
	err = api.menuService.DeleteMenus(menuIds)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(nil)
}

// GetMenuList 获取菜单列表
func (api *MenuApi) GetMenuList(c *ginplus.Context) {
	var req dto.GetMenuListDTO
	if c.BindQueryPage(&req) {
		return
	}
	var menuVOList []*vo.MenuVO

	// 调用服务获取菜单列表
	menuList, total, err := api.menuService.GetMenuList(&req)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	menuVOList = api.menuService.ConvertMenuVO(menuList)
	c.SuccessWithPage(total, menuVOList)
}

// GetMenuTree 获取菜单树
func (api *MenuApi) GetMenuTree(c *ginplus.Context) {
	menuAll, err := api.menuService.GetNoConstantMenuAll()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	// 构建菜单树
	menuAll = api.menuService.BuildMenuTree(menuAll)

	// 将MenuVO列表转换为MenuTreeVO列表
	var rsp []*vo.MenuTreeVO
	rsp = api.menuService.ConvertMenuTreeVO(menuAll)
	c.Success(rsp)
}

// GetMenuInfo 获取菜单详情
func (api *MenuApi) GetMenuInfo(c *ginplus.Context) {
	// 从URL参数中获取菜单ID
	menuIdStr := c.Query("id")
	if menuIdStr == "" {
		c.FailParams("菜单ID不能为空")
		return
	}

	menuId, err := strconv.Atoi(menuIdStr)
	if err != nil {
		c.FailParams("菜单ID格式错误")
		return
	}

	// 查询菜单详情
	menu, err := bo.MenuModel.
		Where(bo.MenuModel.ID.Eq(menuId)).
		Preload(bo.MenuModel.Buttons).
		First()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	if menu == nil {
		c.Fail("菜单不存在")
		return
	}

	// 转换为VO
	menuVO := &vo.MenuVO{
		Id:              menu.ID,
		ParentId:        menu.ParentId,
		MenuType:        menu.MenuType,
		MenuName:        menu.MenuName,
		RouteName:       menu.RouteName,
		RoutePath:       menu.RoutePath,
		Component:       menu.Component,
		I18NKey:         menu.I18NKey,
		IconType:        menu.IconType,
		Icon:            menu.Icon,
		Status:          menu.Status,
		KeepAlive:       menu.KeepAlive,
		Constant:        menu.Constant,
		Order:           menu.OrderBy,
		Href:            menu.Href,
		FixedIndexInTab: menu.FixedIndexInTab,
		HideInMenu:      menu.HideInMenu,
		ActiveMenu:      menu.ActiveMenu,
		MultiTab:        menu.MultiTab,
	}

	c.Success(menuVO)
}

// GetAllPages 获取所有页面组件
func (api *MenuApi) GetAllPages(c *ginplus.Context) {
	// 获取所有菜单类型为 2菜单 的页面组件 name
	menus, err := bo.MenuModel.
		Where(bo.MenuModel.MenuType.Eq("2")).
		Where(bo.MenuModel.Status.Eq("1")).
		Find()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 提取组件名称
	var rsp []string
	for _, menu := range menus {
		if menu.Component != "" {
			rsp = append(rsp, menu.RouteName)
		}
	}

	c.Success(rsp)
}

// GetMenuIdsByRole 根据角色ID获取菜单ID列表
func (api *MenuApi) GetMenuIdsByRole(c *ginplus.Context) {
	var roleId int
	if c.BindRoleId(&roleId) {
		return
	}
	roleMenuIdsVO, err := api.menuService.GetMenuIdsByRoleId(roleId)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(roleMenuIdsVO)
}

// UpdateMenuIdByRole 更新角色绑定的菜单ID和首页
func (api *MenuApi) UpdateMenuIdByRole(c *ginplus.Context) {
	var roleId int
	var req dto.UpdateMenuIdByRoleDTO
	if c.BindRoleId(&roleId) || c.BindJSON(&req) {
		return
	}
	if err := api.menuService.UpdateMenuIdByRole(roleId, &req); err != nil {
		c.Fail(err.Error())
		return
	}
	c.SuccessMsg("更新角色菜单和首页成功", nil)
}
