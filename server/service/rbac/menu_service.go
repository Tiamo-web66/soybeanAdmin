package rbac

import (
	"errors"
	"log/slog"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/po"
	"soybean-admin-go/model/vo"
	"strings"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

// MenuService 菜单相关服务
type MenuService struct{}

// AddMenu 添加菜单
func (s *MenuService) AddMenu(req *dto.AddMenuDTO, operateId int) (int, error) {
	// 检查父菜单是否存在
	if req.ParentId != 0 {
		exist, err := bo.MenuModel.Where(bo.MenuModel.ID.Eq(req.ParentId)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, errors.New("父菜单不存在")
			}
			return 0, err
		}

		// 确保父菜单存在
		if exist == nil {
			return 0, errors.New("父菜单不存在")
		}
	}

	// 转换为MenuModel
	menu := &po.MenuModel{
		ParentId:        req.ParentId,
		MenuType:        req.MenuType,
		MenuName:        req.MenuName,
		RouteName:       req.RouteName,
		RoutePath:       req.RoutePath,
		Component:       req.Component,
		I18NKey:         req.I18NKey,
		IconType:        req.IconType,
		Icon:            req.Icon,
		Status:          req.Status,
		KeepAlive:       req.KeepAlive,
		Constant:        req.Constant,
		OrderBy:         req.Order,
		Href:            req.Href,
		FixedIndexInTab: req.FixedIndexInTab,
		HideInMenu:      req.HideInMenu,
		ActiveMenu:      req.ActiveMenu,
		MultiTab:        req.MultiTab,
		Query:           req.Query,
		RbacOperatorModel: po.RbacOperatorModel{
			CreateByUserId: operateId,
		},
	}

	// 处理按钮
	if len(req.Buttons) > 0 {
		buttons := make([]*po.ButtonModel, 0, len(req.Buttons))
		for _, b := range req.Buttons {
			buttons = append(buttons, &po.ButtonModel{
				Label: b.Desc,
				Code:  b.Code,
			})
		}
		menu.Buttons = buttons
	}

	// 创建菜单
	err := bo.MenuModel.Create(menu)
	if err != nil {
		return 0, err
	}

	return menu.ID, nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(req *dto.EditMenuDTO, operateId int) error {
	// 检查菜单是否存在
	menu, err := bo.MenuModel.Where(bo.MenuModel.ID.Eq(req.Id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return err
	}

	// 如果修改了父级菜单，检查父菜单是否存在
	if req.ParentId != 0 && req.ParentId != menu.ParentId {
		parentMenu, err := bo.MenuModel.Where(bo.MenuModel.ID.Eq(req.ParentId)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("父菜单不存在")
			}
			return err
		}

		// 确保父菜单存在
		if parentMenu == nil {
			return errors.New("父菜单不存在")
		}
	}

	// 使用事务处理更新
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 使用UpdateSimple来确保所有字段都被更新
		_, err := tx.MenuModel.
			Where(tx.MenuModel.ID.Eq(req.Id)).
			UpdateSimple(
				tx.MenuModel.ParentId.Value(req.ParentId),
				tx.MenuModel.MenuType.Value(req.MenuType),
				tx.MenuModel.MenuName.Value(req.MenuName),
				tx.MenuModel.RouteName.Value(req.RouteName),
				tx.MenuModel.RoutePath.Value(req.RoutePath),
				tx.MenuModel.Component.Value(req.Component),
				tx.MenuModel.I18NKey.Value(req.I18NKey),
				tx.MenuModel.IconType.Value(req.IconType),
				tx.MenuModel.Icon.Value(req.Icon),
				tx.MenuModel.Status.Value(req.Status),
				tx.MenuModel.KeepAlive.Value(req.KeepAlive),
				tx.MenuModel.Constant.Value(req.Constant),
				tx.MenuModel.OrderBy.Value(req.Order),
				tx.MenuModel.Href.Value(req.Href),
				tx.MenuModel.FixedIndexInTab.Value(req.FixedIndexInTab),
				tx.MenuModel.HideInMenu.Value(req.HideInMenu),
				tx.MenuModel.ActiveMenu.Value(req.ActiveMenu),
				tx.MenuModel.MultiTab.Value(req.MultiTab),
				tx.MenuModel.UpdateByUserId.Value(operateId),
				tx.MenuModel.Query.Value(req.Query),
			)
		if err != nil {
			return err
		}

		// 处理按钮（如果有）
		if len(req.Buttons) > 0 {
			// 获取现有的按钮
			existingButtons, err := tx.ButtonModel.
				Where(tx.ButtonModel.MenuId.Eq(menu.ID)).
				Find()
			if err != nil {
				return err
			}

			// 创建按钮代码到按钮对象的映射
			existingButtonMap := make(map[string]*po.ButtonModel)
			for _, b := range existingButtons {
				existingButtonMap[b.Code] = b
			}

			// 需要添加的按钮
			var buttonsToAdd []*po.ButtonModel
			// 需要更新的按钮
			var buttonsToUpdate []*po.ButtonModel
			// 要保留的按钮代码
			buttonCodesToKeep := make(map[string]bool)

			// 处理每个按钮
			for _, b := range req.Buttons {
				buttonCodesToKeep[b.Code] = true
				if existingButton, ok := existingButtonMap[b.Code]; ok {
					// 如果按钮已存在，更新它
					existingButton.Label = b.Desc
					buttonsToUpdate = append(buttonsToUpdate, existingButton)
				} else {
					// 如果按钮不存在，添加它
					buttonsToAdd = append(buttonsToAdd, &po.ButtonModel{
						MenuId: menu.ID,
						Label:  b.Desc,
						Code:   b.Code,
					})
				}
			}

			// 要删除的按钮ID
			var buttonIdsToRemove []int
			for _, b := range existingButtons {
				if !buttonCodesToKeep[b.Code] {
					buttonIdsToRemove = append(buttonIdsToRemove, b.ID)
				}
			}

			// 删除不再需要的按钮
			if len(buttonIdsToRemove) > 0 {
				_, err := tx.ButtonModel.
					Where(tx.ButtonModel.ID.In(buttonIdsToRemove...)).
					Delete()
				if err != nil {
					return err
				}
			}

			// 更新现有按钮
			for _, button := range buttonsToUpdate {
				_, err := tx.ButtonModel.
					Where(tx.ButtonModel.ID.Eq(button.ID)).
					Updates(button)
				if err != nil {
					return err
				}
			}

			// 添加新按钮
			for _, button := range buttonsToAdd {
				// 首先检查是否存在相同code的按钮（可能属于其他菜单）
				otherButton, err := tx.ButtonModel.
					Where(tx.ButtonModel.Code.Eq(button.Code)).
					First()

				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						// 如果按钮不存在，创建它
						if err := tx.ButtonModel.Create(button); err != nil {
							return err
						}
					} else {
						return err
					}
				} else {
					// 如果按钮存在但属于其他菜单，更新它的菜单ID和标签
					otherButton.MenuId = button.MenuId
					otherButton.Label = button.Label
					_, err := tx.ButtonModel.
						Where(tx.ButtonModel.ID.Eq(otherButton.ID)).
						Updates(otherButton)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(id int) error {
	// 检查菜单是否存在
	menu, err := bo.MenuModel.Where(bo.MenuModel.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return err
	}

	// 确保菜单存在
	if menu == nil {
		return errors.New("菜单不存在")
	}

	// 检查是否有子菜单
	count, err := bo.MenuModel.Where(bo.MenuModel.ParentId.Eq(id)).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请先删除子菜单")
	}

	// 使用事务删除菜单及相关数据
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 删除菜单关联的按钮
		_, err := tx.ButtonModel.Where(tx.ButtonModel.MenuId.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// 删除菜单关联的API
		_, err = tx.ApiModel.Where(tx.ApiModel.MenuId.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// 删除菜单与角色的关联 - 使用原生SQL
		sql := "DELETE FROM l_role_menus WHERE menu_model_id = ?"
		err = tx.MenuModel.UnderlyingDB().Exec(sql, id).Error
		if err != nil {
			return err
		}

		// 删除菜单
		_, err = tx.MenuModel.Where(tx.MenuModel.ID.Eq(id)).Delete()
		return err
	})
}

// DeleteMenus 批量删除菜单
func (s *MenuService) DeleteMenus(ids []int) error {
	if len(ids) == 0 {
		return errors.New("未选择要删除的菜单")
	}

	// 检查是否有子菜单
	count, err := bo.MenuModel.Where(bo.MenuModel.ParentId.In(ids...)).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请先删除子菜单")
	}

	// 使用事务批量删除
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 将int切片转换为uint切片
		uintIds := make([]int, len(ids))
		for i, id := range ids {
			uintIds[i] = id
		}

		// 删除菜单关联的按钮
		_, err := tx.ButtonModel.Where(tx.ButtonModel.MenuId.In(uintIds...)).Delete()
		if err != nil {
			return err
		}

		// 删除菜单关联的API
		_, err = tx.ApiModel.Where(tx.ApiModel.MenuId.In(uintIds...)).Delete()
		if err != nil {
			return err
		}

		// 删除菜单与角色的关联 - 使用原生SQL
		sql := "DELETE FROM l_role_menus WHERE menu_model_id IN (?)"
		err = tx.MenuModel.UnderlyingDB().Exec(sql, ids).Error
		if err != nil {
			return err
		}

		// 删除菜单
		_, err = tx.MenuModel.Where(tx.MenuModel.ID.In(ids...)).Delete()
		return err
	})
}

// GetMenuList 获取菜单列表,一级菜单 & 非常量菜单
func (s *MenuService) GetMenuList(req *dto.GetMenuListDTO) ([]*po.MenuModel, int64, error) {
	u := bo.MenuModel
	// 分页查询
	menuQuery := u.
		Where(u.Constant.Value(false)).
		Where(u.ParentId.Eq(0))

	// 计算总数
	total, err := menuQuery.Count()
	if err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	menus, err := menuQuery.
		Preload(u.Buttons).
		Preload(u.CreatedBy).
		Preload(u.UpdatedBy).
		Order(u.OrderBy).
		Scopes(req.Scope()).
		Find()
	if err != nil {
		return nil, 0, err
	}
	for i := range menus {
		err = s.loadChildren(menus[i])
		if err != nil {
			return nil, 0, err
		}
	}
	return menus, total, nil
}

// loadChildren 递归加载子菜单,虽然这种方式效率低下,但是因为是分页的,所以暂时不考虑优化
func (s *MenuService) loadChildren(menu *po.MenuModel) error {
	u := bo.MenuModel
	children, err := u.Where(u.ParentId.Eq(menu.ID)).Find()
	if err != nil {
		return err
	}
	menu.Children = children
	for i := range children {
		err = s.loadChildren(menu.Children[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMenuAll 获取所有菜单列表
// constant 是否常量路由
func (s *MenuService) GetMenuAll(constant bool) ([]*po.MenuModel, error) {
	u := bo.MenuModel
	// 查询数据
	menus, err := u.Where(u.Constant.Value(constant)).
		Preload(u.Buttons).
		Preload(u.CreatedBy).
		Preload(u.UpdatedBy).
		Order(u.OrderBy).
		Find()
	if err != nil {
		return nil, err
	}

	return menus, nil
}

// GetNoConstantMenuAll 获取所有非常量菜单列表
// 注意,这个函数没有关联查询菜单的按钮和API和CreateBy UpdateBy
func (s *MenuService) GetNoConstantMenuAll() ([]*po.MenuModel, error) {
	u := bo.MenuModel
	// 查询数据
	menus, err := u.Where(u.Constant.Value(false)).
		Order(u.OrderBy).
		Find()
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// ConvertFrontMenu 将菜单列表转换为前端路由
func (s *MenuService) ConvertFrontMenu(menus []*po.MenuModel) []*vo.FrontRouteItemVO {
	if len(menus) == 0 {
		return []*vo.FrontRouteItemVO{}
	}

	result := make([]*vo.FrontRouteItemVO, 0, len(menus))
	for _, menu := range menus {

		// 创建前端路由对象
		frontRoute := &vo.FrontRouteItemVO{
			Name:      menu.RouteName,
			Path:      menu.RoutePath,
			Component: menu.Component,
			// 设置Meta信息
			Meta: vo.FrontRouteMetaVO{
				Title:      menu.MenuName,
				I18NKey:    menu.I18NKey,
				Icon:       menu.Icon,
				Order:      menu.OrderBy,
				HideInMenu: menu.HideInMenu,
				MultiTab:   menu.MultiTab,
				ActiveMenu: menu.ActiveMenu,
				KeepAlive:  menu.KeepAlive,
				Constant:   menu.Constant,
			},
		}
		result = append(result, frontRoute)

		// 处理props属性 - 判断路径中是否包含":"参数
		if strings.Contains(menu.RoutePath, ":") {
			frontRoute.Props = true
		}

		// 如果有子菜单但没有组件，设置重定向到第一个子菜单
		if menu.Component == "" && len(menu.Children) > 0 {
			frontRoute.Redirect = menu.Children[0].RoutePath
		}
		// 递归处理子菜单
		frontRoute.Children = s.ConvertFrontMenu(menu.Children)
	}

	return result
}

// ConvertMenuVO 将菜单列表转换为MenuVO
func (s *MenuService) ConvertMenuVO(menus []*po.MenuModel) []*vo.MenuVO {
	if len(menus) == 0 {
		return []*vo.MenuVO{}
	}

	result := make([]*vo.MenuVO, 0, len(menus))
	for _, menu := range menus {
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
			CreateBy:        menu.SafeCreateByNickName(),
			UpdateBy:        menu.SafeUpdateByNickName(),
			CreateTime:      menu.CreateTime(),
			UpdateTime:      menu.UpdateTime(),
			Query:           menu.Query,
			Buttons:         s.ConvertButtonVO(menu.Buttons),
		}

		// 递归处理子菜单
		if menu.Children != nil && len(menu.Children) > 0 {
			menuVO.Children = s.ConvertMenuVO(menu.Children)
		}

		result = append(result, menuVO)
	}

	return result
}

// ConvertButtonVO []*ButtonModel 转 []*ButtonVO
func (s *MenuService) ConvertButtonVO(buttons []*po.ButtonModel) []*vo.ButtonVO {
	if len(buttons) == 0 {
		return []*vo.ButtonVO{}
	}
	result := make([]*vo.ButtonVO, 0, len(buttons))
	for _, button := range buttons {
		buttonVO := &vo.ButtonVO{
			Code: button.Code,
			Desc: button.Label,
		}
		result = append(result, buttonVO)
	}
	return result
}

// ConvertMenuTreeVO 将菜单列表转换为MenuTreeVO
func (s *MenuService) ConvertMenuTreeVO(menus []*po.MenuModel) []*vo.MenuTreeVO {
	if len(menus) == 0 {
		return []*vo.MenuTreeVO{}
	}

	result := make([]*vo.MenuTreeVO, 0, len(menus))
	for _, menu := range menus {
		treeVO := &vo.MenuTreeVO{
			Id:    menu.ID,
			Label: menu.MenuName,
			PId:   menu.ParentId,
		}

		// 递归处理子菜单
		if menu.Children != nil && len(menu.Children) > 0 {
			treeVO.Children = s.ConvertMenuTreeVO(menu.Children)
		}

		result = append(result, treeVO)
	}

	return result
}

// GetMenuHomeByRoleIds 根据角色ID获取菜单首页
func (s *MenuService) GetMenuHomeByRoleIds(isSupper bool, roleIds ...int) ([]*po.MenuModel, error) {
	u := bo.RoleModel
	if isSupper {
		roleIds = []int{1}
	}
	roles, err := u.
		Where(u.HomeMenuId.In(roleIds...)).
		Preload(u.HomeMenu).
		Order(u.ID).
		Find()
	if err != nil {
		slog.Error("GetMenuHomeByRoleIds", "err", err)
		return nil, err
	}
	if len(roles) == 0 {
		return nil, nil
	}
	var menus []*po.MenuModel
	for _, role := range roles {
		if role.HomeMenu != nil {
			menus = append(menus, role.HomeMenu)
		}
	}
	return menus, err
}

// GetMenuListByRoleIds 根据角色ID获取非常量菜单列表,
// 注意,这个函数没有关联查询菜单的按钮和API和CreateBy UpdateBy
func (s *MenuService) GetMenuListByRoleIds(isSupper bool, roleIds ...int) ([]*po.MenuModel, error) {
	u := bo.MenuModel
	var menus []*po.MenuModel
	var queryErr error
	if isSupper {
		// 超级管理员获取所有启用的菜单
		menus, queryErr = u.
			Where(u.Constant.Value(false)).
			Order(bo.MenuModel.OrderBy).
			Find()
	} else {
		if len(roleIds) == 0 {
			return []*po.MenuModel{}, nil
		}
		// 获取角色关联的菜单
		roles, err := bo.RoleModel.
			Where(bo.RoleModel.ID.In(roleIds...)).
			Preload(bo.RoleModel.Menus).
			Find()
		if err != nil {
			slog.Error("GetMenuListByRoleIds", "err", err)
			return nil, err
		}
		// 使用 lo.Uniq 提取唯一的菜单 ID
		menuIds := lo.Uniq(lo.FlatMap(roles, func(role *po.RoleModel, _ int) []int {
			return lo.Map(role.Menus, func(menu *po.MenuModel, _ int) int {
				return menu.ID
			})
		}))
		if len(menuIds) > 0 {
			menus, queryErr = bo.MenuModel.
				Where(u.Constant.Value(false)).
				Where(bo.MenuModel.ID.In(menuIds...)).
				Where(bo.MenuModel.Status.Eq(po.MenuStatusEnable)).
				Order(bo.MenuModel.OrderBy).
				Find()
		}
	}
	if queryErr != nil {
		slog.Error("GetMenuListByRoleIds", "err", queryErr)
		return nil, queryErr
	}
	return s.BuildMenuTree(menus), nil
}

// BuildMenuTree 构建菜单树
func (s *MenuService) BuildMenuTree(menus []*po.MenuModel) []*po.MenuModel {
	// 使用lo将slice转为map, key为id, value为menu
	menuMap := lo.SliceToMap(menus, func(menu *po.MenuModel) (int, *po.MenuModel) {
		return menu.ID, menu
	})

	// 根菜单
	var rootMenus []*po.MenuModel

	// 构建树
	for _, menu := range menuMap {
		if menu.ParentId == 0 {
			// 根菜单
			rootMenus = append(rootMenus, menu)
		} else {
			// 子菜单，添加到父菜单的children中
			// 这里为什么不用判断parent是否为nil,这就需要需要前端配合,或者在给role绑定menu的时候,不允许单独绑子菜单,除非绑定父菜单
			if parent, ok := menuMap[menu.ParentId]; ok {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	// 如果没有根菜单，则直接返回原始菜单列表
	if len(rootMenus) == 0 {
		return menus
	}

	return rootMenus
}

// GetAllButtons 获取所有按钮
func (s *MenuService) GetAllButtons() ([]*po.ButtonModel, error) {
	return bo.ButtonModel.Find()
}

// GetButtonIdsByRole 获取角色拥有的按钮ID
func (s *MenuService) GetButtonIdsByRole(roleId int) ([]int, error) {
	role, err := bo.RoleModel.Where(bo.RoleModel.ID.Eq(roleId)).Preload(bo.RoleModel.Buttons).First()
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}
	//  超级管理员获取所有按钮
	if po.IsSuper(role) {
		allButtons, err := bo.ButtonModel.Find()
		if err != nil {
			return nil, err
		}
		role.Buttons = allButtons
	}

	ids := lo.Map(role.Buttons, func(btn *po.ButtonModel, _ int) int {
		return btn.ID
	})
	return ids, nil
}

// AssignRoleButtons 给角色重新分配按钮权限
func (s *MenuService) AssignRoleButtons(roleId int, buttonIds []int) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		role, err := tx.RoleModel.Where(tx.RoleModel.ID.Eq(roleId)).Preload(tx.RoleModel.Buttons).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("角色不存在")
			}
			return err
		}
		if po.IsSuper(role) {
			return errors.New("不允许修改超级管理员角色")
		}
		var buttons []*po.ButtonModel
		if len(buttonIds) > 0 {
			buttons, err = tx.ButtonModel.Where(tx.ButtonModel.ID.In(buttonIds...)).Find()
			if err != nil {
				return err
			}
			if len(buttons) != len(buttonIds) {
				return errors.New("部分按钮不存在")
			}
		}
		if err := tx.RoleModel.Buttons.Model(role).Replace(buttons...); err != nil {
			return err
		}
		return nil
	})
}

// GetMenuIdsByRoleId 获取角色关联的菜单ID
func (s *MenuService) GetMenuIdsByRoleId(roleId int) (*vo.RoleMenuIdsVO, error) {
	// 查询角色关联的菜单
	role, err := bo.RoleModel.
		Where(bo.RoleModel.ID.Eq(roleId)).
		Preload(bo.RoleModel.Menus).
		Preload(bo.RoleModel.HomeMenu).
		Take()
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}
	// 如果是超级管理员，则返回所有菜单
	if po.IsSuper(role) {
		allMenus, err := bo.MenuModel.Find()
		if err != nil {
			return nil, err
		}
		role.Menus = allMenus
	}

	menuIDs := make([]int, 0)
	for _, menu := range role.Menus {
		menuIDs = append(menuIDs, menu.ID)
	}
	routeName := "home"
	if role.HomeMenu != nil {
		routeName = role.HomeMenu.RouteName
	}
	return &vo.RoleMenuIdsVO{
		MenuIds:       menuIDs,
		HomeRouteName: routeName,
	}, nil
}

// UpdateMenuIdByRole 更新角色绑定的菜单ID和首页
func (s *MenuService) UpdateMenuIdByRole(roleId int, req *dto.UpdateMenuIdByRoleDTO) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询角色
		role, err := tx.RoleModel.Where(tx.RoleModel.ID.Eq(roleId)).Preload(tx.RoleModel.Menus).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("角色不存在")
			}
			return err
		}
		// 检查是否是超级管理员,超级管理员不能修改菜单
		if po.IsSuper(role) {
			return errors.New("超级管理员不能修改菜单")
		}

		// 查询所有菜单
		var menus []*po.MenuModel
		if len(req.MenuIds) > 0 {
			menus, err = tx.MenuModel.Where(tx.MenuModel.ID.In(req.MenuIds...)).Find()
			if err != nil {
				return err
			}
			if len(menus) != len(req.MenuIds) {
				return errors.New("部分菜单不存在")
			}
		}
		// 替换角色菜单关联
		if err := tx.RoleModel.Menus.Model(role).Replace(menus...); err != nil {
			return err
		}
		// 更新首页
		if req.HomeRouteName != "" {
			// 查找首页菜单
			var homeMenu *po.MenuModel
			homeMenu, err = tx.MenuModel.Where(tx.MenuModel.RouteName.Eq(req.HomeRouteName)).First()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("首页菜单不存在")
				}
				return err
			}
			_, err = tx.RoleModel.Where(tx.RoleModel.ID.Eq(roleId)).UpdateSimple(tx.RoleModel.HomeMenuId.Value(homeMenu.ID))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
