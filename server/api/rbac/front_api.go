package rbac

import (
	"github.com/samber/lo"
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/po"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
	"soybean-admin-go/service/rbac"
)

// FrontAPI 前端路由API控制器
type FrontAPI struct {
	menuService rbac.MenuService
}

// GetUserRoutes 获取用户路由
func (api *FrontAPI) GetUserRoutes(c *ginplus.Context) {
	var userID int
	if exists := c.GetOperateId(&userID); !exists {
		return
	}

	// 查询用户信息及其角色
	currentUser, err := bo.UserModel.WithContext(global.Context()).
		Where(bo.UserModel.ID.Eq(userID)).
		Preload(bo.UserModel.Roles).
		First()
	if err != nil {
		c.Fail(err.Error())
		return
	}

	if currentUser == nil {
		c.Fail("用户不存在")
		return
	}

	// 检查用户角色
	if len(currentUser.Roles) == 0 {
		c.Success(vo.FrontRouteVO{Routes: []*vo.FrontRouteItemVO{}})
		return
	}

	// 提取角色ID
	var roleIds = lo.Map(currentUser.Roles, func(role *po.RoleModel, _ int) int { return role.ID })
	var isSuperAdmin = po.IsSuper(currentUser.Roles...)
	// 获取菜单
	menus, err := api.menuService.GetMenuListByRoleIds(isSuperAdmin, roleIds...)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	// 获取首页菜单
	homeMenus, err := api.menuService.GetMenuHomeByRoleIds(isSuperAdmin, roleIds...)
	if err != nil {
		c.Fail(err.Error())
		return
	}

	// 构建前端路由
	var rsp vo.FrontRouteVO
	rsp.Routes = api.menuService.ConvertFrontMenu(menus)
	if len(homeMenus) > 0 {
		rsp.Home = homeMenus[0].RouteName
	} else {
		// 找到第一个没有子菜单的菜单作为主页
		for _, route := range rsp.Routes {
			if len(route.Children) == 0 {
				rsp.Home = route.Name
				break
			}
		}
		// 如果没有找到主页，则使用第一个路由作为主页
		if rsp.Home == "" && len(rsp.Routes) > 0 {
			rsp.Home = rsp.Routes[0].Name
		}
	}
	c.Success(rsp)
}

// Exist 路由时候存在
func (api *FrontAPI) Exist(c *ginplus.Context) {
	//从参数获取routeName参数
	routeName := c.Query("routeName")
	if routeName == "" {
		c.Success(false)
		return
	}

	// 查询路由是否存在
	count, err := bo.MenuModel.
		Where(bo.MenuModel.RouteName.Eq(routeName)).
		Where(bo.MenuModel.Status.Eq(po.MenuStatusEnable)).
		Count()
	if err != nil {
		c.Fail(err.Error())
		return
	}
	c.Success(count > 0)
}

// GetFixedRoutes 获取固定的路由数据(也就是常量路由,常量路由不会有子路由,暂且这么说)
func (api *FrontAPI) GetFixedRoutes(c *ginplus.Context) {
	menuAll, err := api.menuService.GetMenuAll(true)
	if err != nil {
		c.Fail(err.Error())
		return
	}
	var rsp []*vo.FrontRouteItemVO
	rsp = api.menuService.ConvertFrontMenu(menuAll)
	c.Success(rsp)
}
