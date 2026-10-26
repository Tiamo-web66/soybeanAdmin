package router

import (
	"soybean-admin-go/api/rbac"
	"soybean-admin-go/middleware"
	"soybean-admin-go/pkg/ginplus"
)

// InitSystemManageRouter 初始化系统管理路由
func InitSystemManageRouter(r *ginplus.RouterGroup) {
	menuApi := rbac.MenuApi{}
	roleApi := rbac.RoleAPI{}
	userApi := rbac.UserAPI{}
	buttonApi := rbac.ButtonApi{}
	apiApi := rbac.ApiApi{}

	// 先创建Group，然后再使用中间件
	systemManageRouterBase := r.Group("/systemManage")

	// 需要JWT验证的路由
	systemManageRouter := systemManageRouterBase.Group("")
	systemManageRouter.Use(middleware.JWTAuth).RequireAuth()
	{
		// 查询接口
		systemManageRouter.GET("/getAllPages", "获取所有页面", "获取系统所有页面组件信息", menuApi.GetAllPages)
		systemManageRouter.GET("/getMenuTree", "获取菜单树", "获取系统菜单树形结构", menuApi.GetMenuTree)
		systemManageRouter.GET("/getAllButton", "获取所有按钮", "获取系统所有按钮信息", buttonApi.GetAllButtons)
		systemManageRouter.GET("/getAllApi", "获取所有API", "获取系统所有API信息", apiApi.GetAllApis)
	}

	// 需要权限验证的路由
	systemManageRouterWithAuth := systemManageRouterBase.Group("")
	systemManageRouterWithAuth.Use(middleware.AuthCheck).RequireAuth()
	{
		// 菜单管理接口
		systemManageRouterWithAuth.GET("/getMenuList/v2", "获取菜单列表", "获取系统菜单列表，支持分页", menuApi.GetMenuList)
		systemManageRouterWithAuth.POST("/addMenu", "添加菜单", "添加新的系统菜单", menuApi.AddMenu)
		systemManageRouterWithAuth.PUT("/updateMenu", "更新菜单", "更新已有的系统菜单", menuApi.UpdateMenu)
		systemManageRouterWithAuth.DELETE("/deleteMenu", "删除菜单", "删除单个系统菜单", menuApi.DeleteMenu)
		systemManageRouterWithAuth.DELETE("/deleteMenus", "批量删除菜单", "批量删除多个系统菜单", menuApi.DeleteMenus)
		systemManageRouterWithAuth.GET("/menuInfo/:id", "获取菜单详情", "获取指定ID的菜单详细信息", menuApi.GetMenuInfo)

		// 用户管理接口
		systemManageRouterWithAuth.GET("/getUserList", "获取用户列表", "获取系统用户列表，支持分页", userApi.GetUserList)
		systemManageRouterWithAuth.POST("/addUser", "创建用户", "创建新的系统用户", userApi.CreateUser)
		systemManageRouterWithAuth.PUT("/updateUser", "更新用户", "更新已有的系统用户", userApi.UpdateUser)
		systemManageRouterWithAuth.DELETE("/deleteUser", "删除用户", "删除单个系统用户", userApi.DeleteUser)
		systemManageRouterWithAuth.DELETE("/deleteUsers", "批量删除用户", "批量删除多个系统用户", userApi.DeleteUsers)

		// 角色管理接口
		systemManageRouterWithAuth.GET("/getRoleList", "获取角色列表", "获取系统角色列表，支持分页", roleApi.GetRoleList)
		systemManageRouterWithAuth.GET("/getAllRoles", "获取所有角色", "获取系统所有角色信息", roleApi.GetAllRoles)
		systemManageRouterWithAuth.POST("/addRole", "创建角色", "创建新的系统角色", roleApi.CreateRole)
		systemManageRouterWithAuth.PUT("/updateRole", "更新角色", "更新已有的系统角色", roleApi.UpdateRole)
		systemManageRouterWithAuth.DELETE("/deleteRole", "删除角色", "删除系统角色", roleApi.DeleteRole)
		systemManageRouterWithAuth.DELETE("/deleteRoles", "批量删除角色", "批量删除多个系统角色", roleApi.DeleteRoles)
		//角色 的菜单管理
		systemManageRouterWithAuth.GET("/role/:roleId/menus", "获取角色的菜单ID列表", "获取指定角色的菜单ID列表", menuApi.GetMenuIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/menus", "给角色分配菜单权限", "给角色重新分配菜单权限", menuApi.UpdateMenuIdByRole)
		//角色 的按钮管理
		systemManageRouterWithAuth.GET("/role/:roleId/buttons", "获取角色的按钮ID列表", "获取指定角色的按钮ID列表", buttonApi.GetButtonIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/buttons", "给角色分配按钮权限", "给角色重新分配按钮权限", buttonApi.UpdateButtonIdsByRole)
		//角色 的API管理
		systemManageRouterWithAuth.GET("/role/:roleId/apis", "获取角色的API ID列表", "获取指定角色的API ID列表", apiApi.GetApiIdsByRole)
		systemManageRouterWithAuth.PUT("/role/:roleId/apis", "给角色分配API权限", "给角色重新分配API权限", apiApi.UpdateApiIdsByRole)
	}
}
