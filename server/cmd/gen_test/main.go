package main

import (
	"errors"
	"fmt"
	"os"
	"soybean-admin-go/global"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/po"
	"soybean-admin-go/utils/password"
	"time"

	"gorm.io/gorm"
)

// addSuperAdmin 创建超管角色和超管用户并绑定,超管角色不需要绑定菜单之类的,在业务操作的时候,超管角色直接查出所有
func addSuperAdmin() (*po.UserModel, *po.RoleModel, error) {
	q := bo.Q
	// 哈希密码
	hashedPassword, err := password.HashPassword("123456")
	if err != nil {
		return nil, nil, fmt.Errorf("哈希密码失败: %w", err)
	}

	nowTime := time.Now().UnixMilli()

	// 使用gorm创建或更新超管用户 (ID=1)
	superUser := &po.UserModel{
		Model: po.Model{
			ID:        1,
			CreatedAt: nowTime,
			UpdatedAt: nowTime,
		},
		UserName:   "Soybean",
		Password:   hashedPassword,
		NickName:   "超级管理员",
		UserGender: po.UserGenderMan,
		UserPhone:  "12345678901",
		UserEmail:  "admin@example.com",
		Status:     po.UserStatusEnable,
		OperatorModel: po.OperatorModel{
			CreateById: 1,
			UpdateById: 1,
		},
	}

	// 使用Upsert功能创建或更新超管用户
	err = q.UserModel.Save(superUser)
	if err != nil {
		return nil, nil, fmt.Errorf("upsert超管用户失败: %w", err)
	}

	// 设置序列值确保下一个ID是2
	if err := global.DB.Exec("SELECT setval('l_users_id_seq', 1, true)").Error; err != nil {
		return nil, nil, fmt.Errorf("设置序列当前值失败: %w", err)
	}

	// 查询创建的用户
	superUser, err = q.UserModel.Where(q.UserModel.ID.Eq(1)).First()
	if err != nil {
		return nil, nil, fmt.Errorf("查询超管用户失败: %w", err)
	}

	// 创建或更新超管角色
	superRole := &po.RoleModel{
		Model: po.Model{
			CreatedAt: nowTime,
			UpdatedAt: nowTime,
		},
		Name:       "超级管理员",
		Code:       po.SuperRoleCode,
		Desc:       "系统超级管理员，拥有所有权限",
		Status:     po.UserStatusEnable,
		HomeMenuId: 1,
		RbacOperatorModel: po.RbacOperatorModel{
			CreateByUserId: 1,
			UpdateByUserId: 1,
		},
	}

	// 先查询是否已存在该角色
	existingRole, err := q.RoleModel.Where(q.RoleModel.Code.Eq(po.SuperRoleCode)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("查询超管角色失败: %w", err)
	}

	// 如果已存在，更新它
	if existingRole != nil {
		superRole.ID = existingRole.ID
		_, err = q.RoleModel.Where(q.RoleModel.ID.Eq(existingRole.ID)).Updates(superRole)
		if err != nil {
			return nil, nil, fmt.Errorf("更新超管角色失败: %w", err)
		}
	} else {
		// 否则创建新角色
		err = q.RoleModel.Create(superRole)
		if err != nil {
			return nil, nil, fmt.Errorf("创建超管角色失败: %w", err)
		}
	}

	// 查询创建的角色
	superRoleResult, err := q.RoleModel.Where(q.RoleModel.Code.Eq(po.SuperRoleCode)).First()
	if err != nil {
		return nil, nil, fmt.Errorf("查询超管角色失败: %w", err)
	}

	// 创建用户角色关联
	// 先检查关联是否已存在
	var userRole struct {
		UserModelID int64
		RoleModelID int64
	}

	result := global.DB.Table("l_user_roles").
		Where("user_model_id = ? AND role_model_id = ?", superUser.ID, superRoleResult.ID).
		First(&userRole)

	// 如果关联不存在，则创建
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = q.UserModel.Roles.Model(superUser).Append(superRoleResult)
		if err != nil {
			return nil, nil, fmt.Errorf("创建用户角色关联失败: %w", err)
		}
	}

	return superUser, superRoleResult, nil
}

// 添加初始数据到数据库
func addInitialData() error {
	q := bo.Q

	//创建超管用户和角色
	superUser, _, err := addSuperAdmin()
	if err != nil {
		return fmt.Errorf("添加超管数据失败: %w", err)
	}

	// 开启事务
	return q.Transaction(func(tx *bo.Query) error {
		// 添加角色数据
		roles := []*po.RoleModel{
			{
				Name:   "管理员",
				Code:   "R_ADMIN",
				Desc:   "系统超级管理员，拥有所有权限",
				Status: po.UserStatusEnable,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID, // 超管用户ID
					UpdateByUserId: superUser.ID, // 超管用户ID
				},
			},
			{
				Name:   "普通用户",
				Code:   "R_USER",
				Desc:   "普通用户，拥有基本权限",
				Status: po.UserStatusEnable,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID, // 超管用户ID
					UpdateByUserId: superUser.ID, // 超管用户ID
				},
			},
		}

		// 添加角色
		if err := tx.RoleModel.Create(roles...); err != nil {
			return fmt.Errorf("创建角色失败: %w", err)
		}

		// 添加菜单
		menus := []*po.MenuModel{
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "首页",
				RouteName:       "home",
				RoutePath:       "/home",
				Component:       "layout.base$view.home",
				I18NKey:         "route.home",
				IconType:        "1",
				Icon:            "mdi:monitor-dashboard",
				Status:          "1",
				KeepAlive:       true,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "1", // 目录
				MenuName:        "系统功能",
				RouteName:       "function",
				RoutePath:       "/function",
				Component:       "layout.base",
				I18NKey:         "route.function",
				IconType:        "1",
				Icon:            "icon-park-outline:all-application",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         3,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        2,   // 功能的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "多标签页",
				RouteName:       "function_multi-tab",
				RoutePath:       "/function/multi-tab",
				Component:       "view.function_multi-tab",
				I18NKey:         "route.function_multi-tab",
				IconType:        "1",
				Icon:            "ic:round-tab",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      true,
				ActiveMenu:      "function_tab",
				MultiTab:        true,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        2,   // 功能的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "标签页",
				RouteName:       "function_tab",
				RoutePath:       "/function/tab",
				Component:       "view.function_tab",
				I18NKey:         "route.function_tab",
				IconType:        "1",
				Icon:            "ic:round-tab",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         2,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "1", // 目录
				MenuName:        "异常页",
				RouteName:       "exception",
				RoutePath:       "/exception",
				Component:       "layout.base",
				I18NKey:         "route.exception",
				IconType:        "1",
				Icon:            "ant-design:exception-outlined",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         3,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        5,   // 异常页的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "403",
				RouteName:       "exception_403",
				RoutePath:       "/exception/403",
				Component:       "view.403",
				I18NKey:         "route.exception_403",
				IconType:        "1",
				Icon:            "ic:baseline-block",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        5,   // 异常页的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "404",
				RouteName:       "exception_404",
				RoutePath:       "/exception/404",
				Component:       "view.404",
				I18NKey:         "route.exception_404",
				IconType:        "1",
				Icon:            "ic:baseline-web-asset-off",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         2,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        5,   // 异常页的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "500",
				RouteName:       "exception_500",
				RoutePath:       "/exception/500",
				Component:       "view.500",
				I18NKey:         "route.exception_500",
				IconType:        "1",
				Icon:            "ic:baseline-wifi-off",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         3,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "1", // 目录
				MenuName:        "多级菜单",
				RouteName:       "multi-menu",
				RoutePath:       "/multi-menu",
				Component:       "layout.base",
				I18NKey:         "route.multi-menu",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         4,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        9,   // 多级菜单的子菜单
				MenuType:        "1", // 目录
				MenuName:        "一级子菜单1",
				RouteName:       "multi-menu_first",
				RoutePath:       "/multi-menu/first",
				Component:       "",
				I18NKey:         "route.multi-menu_first",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        10,  // 一级子菜单1的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "二级子菜单",
				RouteName:       "multi-menu_first_child",
				RoutePath:       "/multi-menu/first/child",
				Component:       "view.multi-menu_first_child",
				I18NKey:         "route.multi-menu_first_child",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        9,   // 多级菜单的子菜单
				MenuType:        "1", // 目录
				MenuName:        "一级子菜单2",
				RouteName:       "multi-menu_second",
				RoutePath:       "/multi-menu/second",
				Component:       "",
				I18NKey:         "route.multi-menu_second",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         13,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        12,  // 一级子菜单2的子菜单
				MenuType:        "1", // 目录
				MenuName:        "二级子菜单2",
				RouteName:       "multi-menu_second_child",
				RoutePath:       "/multi-menu/second/child",
				Component:       "",
				I18NKey:         "route.multi-menu_second_child",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        14,  // 二级子菜单2的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "三级菜单",
				RouteName:       "multi-menu_second_child_home",
				RoutePath:       "/multi-menu/second/child/home",
				Component:       "view.multi-menu_second_child_home",
				I18NKey:         "route.multi-menu_second_child_home",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "1", // 目录
				MenuName:        "系统管理",
				RouteName:       "manage",
				RoutePath:       "/manage",
				Component:       "layout.base",
				I18NKey:         "route.manage",
				IconType:        "1",
				Icon:            "carbon:cloud-service-management",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         5,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        16,  // 系统管理的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "用户管理",
				RouteName:       "manage_user",
				RoutePath:       "/manage/user",
				Component:       "view.manage_user",
				I18NKey:         "route.manage_user",
				IconType:        "1",
				Icon:            "ic:round-manage-accounts",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         1,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        16,  // 系统管理的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "角色管理",
				RouteName:       "manage_role",
				RoutePath:       "/manage/role",
				Component:       "view.manage_role",
				I18NKey:         "route.manage_role",
				IconType:        "1",
				Icon:            "carbon:user-role",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         2,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        16,  // 系统管理的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "菜单管理",
				RouteName:       "manage_menu",
				RoutePath:       "/manage/menu",
				Component:       "view.manage_menu",
				I18NKey:         "route.manage_menu",
				IconType:        "1",
				Icon:            "material-symbols:route",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         3,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        16,  // 系统管理的子菜单
				MenuType:        "2", // 菜单
				MenuName:        "用户详情",
				RouteName:       "manage_user-detail",
				RoutePath:       "/manage/user-detail/:id",
				Component:       "view.manage_user-detail",
				I18NKey:         "route.manage_user-detail",
				IconType:        "",
				Icon:            "",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         4,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      true,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "关于",
				RouteName:       "about",
				RoutePath:       "/about",
				Component:       "layout.base$view.about",
				I18NKey:         "route.about",
				IconType:        "1",
				Icon:            "fluent:book-information-24-regular",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         6,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
		}
		//常量路由
		constantMenus := []*po.MenuModel{
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "登录",
				RouteName:       "login",
				RoutePath:       "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
				Component:       "layout.blank$view.login",
				I18NKey:         "route.login",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         7,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				Constant:        true,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "403",
				RouteName:       "403",
				RoutePath:       "/403",
				Component:       "layout.blank$view.403",
				I18NKey:         "route.403",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         8,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				Constant:        true,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "404",
				RouteName:       "404",
				RoutePath:       "/404",
				Component:       "layout.blank$view.404",
				I18NKey:         "route.404",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         8,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				Constant:        true,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
			{
				ParentId:        0,
				MenuType:        "2", // 菜单
				MenuName:        "500",
				RouteName:       "500",
				RoutePath:       "/500",
				Component:       "layout.blank$view.500",
				I18NKey:         "route.500",
				IconType:        "1",
				Icon:            "mdi:menu",
				Status:          "1",
				KeepAlive:       false,
				OrderBy:         9,
				Href:            "",
				FixedIndexInTab: 0,
				HideInMenu:      false,
				ActiveMenu:      "",
				MultiTab:        false,
				Constant:        true,
				RbacOperatorModel: po.RbacOperatorModel{
					CreateByUserId: superUser.ID,
					UpdateByUserId: superUser.ID,
				},
			},
		}

		if err := tx.MenuModel.Create(append(menus, constantMenus...)...); err != nil {
			return fmt.Errorf("创建菜单失败: %w", err)
		}

		// 查询用户管理菜单
		userMenu, err := tx.MenuModel.Where(tx.MenuModel.MenuName.Eq("用户管理")).First()
		if err != nil {
			return fmt.Errorf("查询用户管理菜单失败: %w", err)
		}

		// 添加按钮
		buttons := []*po.ButtonModel{
			{
				MenuId: userMenu.ID, // 用户管理菜单下的按钮
				Label:  "新增",
				Code:   "user:add",
			},
			{
				MenuId: userMenu.ID,
				Label:  "编辑",
				Code:   "user:edit",
			},
			{
				MenuId: userMenu.ID,
				Label:  "删除",
				Code:   "user:delete",
			},
		}

		if err := tx.ButtonModel.Create(buttons...); err != nil {
			return fmt.Errorf("创建按钮失败: %w", err)
		}

		adminRole, err := tx.RoleModel.Where(tx.RoleModel.Code.Eq("R_ADMIN")).First()
		if err != nil {
			return fmt.Errorf("查询管理员角色失败: %w", err)
		}

		userRole, err := tx.RoleModel.Where(tx.RoleModel.Code.Eq("R_USER")).First()
		if err != nil {
			return fmt.Errorf("查询普通用户角色失败: %w", err)
		}

		// 查询出所有菜单
		allMenus, err := tx.MenuModel.Find()
		if err != nil {
			return fmt.Errorf("查询所有菜单失败: %w", err)
		}

		// 查询出所有按钮
		allButtons, err := tx.ButtonModel.Find()
		if err != nil {
			return fmt.Errorf("查询所有按钮失败: %w", err)
		}

		// 关联所有菜单到管理员角色
		if err := tx.RoleModel.Menus.Model(adminRole).Append(allMenus...); err != nil {
			return fmt.Errorf("关联菜单到管理员角色失败: %w", err)
		}

		// 关联所有按钮到管理员角色
		if err := tx.RoleModel.Buttons.Model(adminRole).Append(allButtons...); err != nil {
			return fmt.Errorf("关联按钮到管理员角色失败: %w", err)
		}

		// 哈希密码
		hashedPassword, err := password.HashPassword("123456")
		if err != nil {
			return fmt.Errorf("哈希密码失败: %w", err)
		}

		// 创建管理员用户和普通用户,并设置角色
		adminUser := &po.UserModel{
			UserName:   "admin",
			Password:   hashedPassword,
			NickName:   "管理员",
			UserGender: po.UserGenderMan,
			UserPhone:  "13800138001",
			UserEmail:  "admin@example.com",
			Status:     po.UserStatusEnable,
			OperatorModel: po.OperatorModel{
				CreateById: superUser.ID,
				UpdateById: superUser.ID,
			},
		}

		normalUser := &po.UserModel{
			UserName:   "user",
			Password:   hashedPassword,
			NickName:   "普通用户",
			UserGender: po.UserGenderMan,
			UserPhone:  "13800138002",
			UserEmail:  "user@example.com",
			Status:     po.UserStatusEnable,
			OperatorModel: po.OperatorModel{
				CreateById: superUser.ID,
				UpdateById: superUser.ID,
			},
		}

		// 创建用户
		if err := tx.UserModel.Create(adminUser, normalUser); err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		// 关联管理员角色到admin用户
		if err := tx.UserModel.Roles.Model(adminUser).Append(adminRole); err != nil {
			return fmt.Errorf("关联管理员角色到admin用户失败: %w", err)
		}

		// 关联普通用户角色到normal用户
		if err := tx.UserModel.Roles.Model(normalUser).Append(userRole); err != nil {
			return fmt.Errorf("关联普通用户角色到normal用户失败: %w", err)
		}

		return nil
	})
}

// 清除所有初始化的数据
func cleanupData() error {
	q := bo.Q

	// 开启事务
	return q.Transaction(func(tx *bo.Query) error {
		// 清除关联表数据
		tables := []string{
			"l_user_roles",   // 用户角色关联
			"l_role_menus",   // 角色菜单关联
			"l_role_apis",    // 角色API关联
			"l_role_buttons", // 角色按钮关联
		}

		for _, table := range tables {
			if err := global.DB.Exec("DELETE FROM ?", gorm.Expr(table)).Error; err != nil {
				return fmt.Errorf("清除%s表失败: %w", table, err)
			}
		}

		// 清除数据并重置序列
		// 用户表
		users, err := tx.UserModel.Unscoped().Find()
		if err != nil {
			return fmt.Errorf("查询所有用户失败: %w", err)
		}
		if len(users) > 0 {
			if _, err := tx.UserModel.Unscoped().Delete(users...); err != nil {
				return fmt.Errorf("清除用户表失败: %w", err)
			}
		}
		if err := global.DB.Exec("ALTER SEQUENCE l_users_id_seq RESTART WITH 1").Error; err != nil {
			return fmt.Errorf("重置用户表自增id失败: %w", err)
		}

		// 角色表
		roles, err := tx.RoleModel.Unscoped().Find()
		if err != nil {
			return fmt.Errorf("查询所有角色失败: %w", err)
		}
		if len(roles) > 0 {
			if _, err := tx.RoleModel.Unscoped().Delete(roles...); err != nil {
				return fmt.Errorf("清除角色表失败: %w", err)
			}
		}
		if err := global.DB.Exec("ALTER SEQUENCE l_role_id_seq RESTART WITH 1").Error; err != nil {
			return fmt.Errorf("重置角色表自增id失败: %w", err)
		}

		// API表
		apis, err := tx.ApiModel.Unscoped().Find()
		if err != nil {
			return fmt.Errorf("查询所有API失败: %w", err)
		}
		if len(apis) > 0 {
			if _, err := tx.ApiModel.Unscoped().Delete(apis...); err != nil {
				return fmt.Errorf("清除API表失败: %w", err)
			}
		}
		if err := global.DB.Exec("ALTER SEQUENCE l_api_id_seq RESTART WITH 1").Error; err != nil {
			return fmt.Errorf("重置API表自增id失败: %w", err)
		}

		// 按钮表
		buttons, err := tx.ButtonModel.Unscoped().Find()
		if err != nil {
			return fmt.Errorf("查询所有按钮失败: %w", err)
		}
		if len(buttons) > 0 {
			if _, err := tx.ButtonModel.Unscoped().Delete(buttons...); err != nil {
				return fmt.Errorf("清除按钮表失败: %w", err)
			}
		}
		if err := global.DB.Exec("ALTER SEQUENCE l_button_id_seq RESTART WITH 1").Error; err != nil {
			return fmt.Errorf("重置按钮表自增id失败: %w", err)
		}

		// 菜单表
		menus, err := tx.MenuModel.Unscoped().Find()
		if err != nil {
			return fmt.Errorf("查询所有菜单失败: %w", err)
		}
		if len(menus) > 0 {
			if _, err := tx.MenuModel.Unscoped().Delete(menus...); err != nil {
				return fmt.Errorf("清除菜单表失败: %w", err)
			}
		}
		if err := global.DB.Exec("ALTER SEQUENCE l_menu_id_seq RESTART WITH 1").Error; err != nil {
			return fmt.Errorf("重置菜单表自增id失败: %w", err)
		}

		return nil
	})
}

func main() {
	// 初始化配置
	err := os.Setenv("CONFIG_PATH", "./config.toml")
	if err != nil {
		fmt.Println("设置环境变量失败：", err)
		return
	}
	initialize.InitAll()

	// 检查数据库连接
	if global.DB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 如果需要清除数据，取消下面的注释
	//if err := cleanupData(); err != nil {
	//	fmt.Println("清除数据失败:", err)
	//	return
	//}
	//fmt.Println("成功清除数据")

	// 如果需要添加初始数据，取消下面的注释
	if err := addInitialData(); err != nil {
		fmt.Println("添加初始数据失败:", err)
		return
	}
	fmt.Println("成功添加初始数据")
}
