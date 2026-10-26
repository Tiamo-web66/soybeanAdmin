package rbac

import (
	"errors"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"log/slog"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/po"
)

// RoleService 角色相关服务
type RoleService struct{}

// CreateRole 创建角色
func (s *RoleService) CreateRole(req *dto.CreateRoleDTO, operatorID int) (int, error) {
	var id int
	err := bo.Q.Transaction(func(tx *bo.Query) error {
		// 检查角色编码是否已存在
		count, err := tx.RoleModel.Where(tx.RoleModel.Code.Eq(req.RoleCode)).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("角色编码已存在")
		}

		// 创建角色
		role := &po.RoleModel{
			Name:   req.RoleName,
			Code:   req.RoleCode,
			Desc:   req.RoleDesc,
			Status: req.Status,
			RbacOperatorModel: po.RbacOperatorModel{
				CreateByUserId: operatorID,
				UpdateByUserId: operatorID,
			},
		}
		err = tx.RoleModel.Create(role)
		if err != nil {
			return err
		}
		id = role.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateRole 重写角色更新逻辑
func (s *RoleService) UpdateRole(req *dto.UpdateRoleDTO, operatorID int) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询角色是否存在
		role, err := tx.RoleModel.Where(tx.RoleModel.ID.Eq(req.Id)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("角色不存在")
			}
			return err
		}
		//不允许更新超管角色
		if po.IsSuper(role) {
			return errors.New("不允许修改超级管理员角色")
		}

		// 如果修改了角色编码，需要检查新编码是否与其他角色冲突
		if req.RoleCode != role.Code {
			count, err := tx.RoleModel.Where(tx.RoleModel.Code.Eq(req.RoleCode), tx.RoleModel.ID.Neq(req.Id)).Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return errors.New("角色编码已被其他角色使用")
			}
		}

		// 更新字段集合
		updateFields := []field.AssignExpr{
			tx.RoleModel.Name.Value(req.RoleName),
			tx.RoleModel.Code.Value(req.RoleCode),
			tx.RoleModel.Desc.Value(req.RoleDesc),
			tx.RoleModel.Status.Value(req.Status),
			tx.RoleModel.UpdateByUserId.Value(operatorID),
		}
		if req.HomeMenuId != nil {
			updateFields = append(updateFields, tx.RoleModel.HomeMenuId.Value(*req.HomeMenuId))
		}
		// 更新角色信息
		_, err = tx.RoleModel.Where(tx.RoleModel.ID.Eq(req.Id)).UpdateSimple(updateFields...)
		if err != nil {
			return err
		}

		// 如果角色编码发生了变化，需要处理缓存
		if req.RoleCode != role.Code {
			if err := initialize.ReloadRoleApis(role.Code); err != nil {
				slog.Error("清除旧角色缓存失败", "roleCode", role.Code, "error", err)
			}
		}

		// 更新角色API缓存
		if err := initialize.ReloadRoleApis(req.RoleCode); err != nil {
			slog.Error("更新角色API缓存失败", "roleCode", req.RoleCode, "error", err)
		}

		return nil
	})
}

// DeleteRole 删除角色（gorm gen重构）
func (s *RoleService) DeleteRole(roleID int) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询角色是否存在
		role, err := tx.RoleModel.Preload(tx.RoleModel.Users).Where(tx.RoleModel.ID.Eq(roleID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("角色不存在")
			}
			return err
		}

		// 不允许删除超级管理员角色
		if po.IsSuper(role) {
			return errors.New("不允许删除超级管理员角色")
		}

		// 检查是否有关联用户
		if len(role.Users) > 0 {
			return errors.New("有用户绑定该角色,无法删除")
		}

		// 删除角色与API的关联
		if err := tx.RoleModel.Apis.Model(role).Clear(); err != nil {
			return err
		}

		// 删除角色与菜单的关联
		if err := tx.RoleModel.Menus.Model(role).Clear(); err != nil {
			return err
		}

		// 删除角色
		if _, err := tx.RoleModel.Delete(role); err != nil {
			return err
		}

		// 清除角色缓存
		if err := initialize.ReloadRoleApis(role.Code); err != nil {
			slog.Error("清除角色缓存失败", "roleCode", role.Code, "error", err)
		}

		return nil
	})
}

// DeleteRoles 批量删除角色
func (s *RoleService) DeleteRoles(roleIDs []int) error {
	if len(roleIDs) == 0 {
		return errors.New("角色ID列表不能为空")
	}
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询所有角色，预加载用户
		roles, err := tx.RoleModel.Preload(tx.RoleModel.Users).Where(tx.RoleModel.ID.In(roleIDs...)).Find()
		if err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return errors.New("部分角色不存在")
		}

		// 检查是否有超级管理员或有关联用户
		for _, role := range roles {
			if po.IsSuper(role) {
				return errors.New("不允许删除超级管理员角色")
			}
			if len(role.Users) > 0 {
				return errors.New("有用户绑定该角色,无法删除")
			}
		}

		// 清理角色与API、菜单的关联
		for _, role := range roles {
			if err := tx.RoleModel.Apis.Model(role).Clear(); err != nil {
				return err
			}
			if err := tx.RoleModel.Menus.Model(role).Clear(); err != nil {
				return err
			}
		}

		// 批量删除角色
		_, err = tx.RoleModel.Where(tx.RoleModel.ID.In(roleIDs...)).Delete()
		if err != nil {
			return err
		}

		// 清除角色缓存
		for _, role := range roles {
			if err := initialize.ReloadRoleApis(role.Code); err != nil {
				slog.Error("清除角色缓存失败", "roleCode", role.Code, "error", err)
			}
		}

		return nil
	})
}
