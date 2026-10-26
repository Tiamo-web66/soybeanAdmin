package initialize

import (
	"context"
	"log/slog"
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/po"
	redisUtil "soybean-admin-go/utils/redis"

	"github.com/samber/lo"
)

// InitRbacCache 初始化RBAC缓存
// 在系统启动时调用，将角色和API权限信息加载到Redis缓存中
func InitRbacCache() {
	ctx := global.Context()

	// 加载所有API到缓存
	if err := loadAllApis(ctx); err != nil {
		slog.Error("加载API到缓存失败", "error", err)
	}

	// 加载所有角色的API权限到缓存
	if err := loadAllRoleApis(ctx); err != nil {
		slog.Error("加载角色API权限到缓存失败", "error", err)
	}

	slog.Info("RBAC缓存初始化完成")
}

// loadAllApis 加载所有API到缓存
func loadAllApis(ctx context.Context) error {
	// 从数据库中查询所有API
	apis, err := bo.ApiModel.Find()
	if err != nil {
		return err
	}

	slog.Info("加载API到缓存", "count", len(apis))

	// 预缓存所有API权限
	for _, api := range apis {
		if err := redisUtil.PreCacheApiAuth(ctx, api); err != nil {
			slog.Error("缓存API权限失败", "api", api.Path, "error", err)
		}
	}

	return nil
}

// loadAllRoleApis 加载所有角色的API权限到缓存
func loadAllRoleApis(ctx context.Context) error {
	// 从数据库中查询所有角色
	roles, err := bo.RoleModel.Preload(bo.RoleModel.Apis).Find()
	if err != nil {
		return err
	}

	slog.Info("加载角色API权限到缓存", "roleCount", len(roles))

	// 缓存每个角色的API权限
	for _, role := range roles {
		if err := redisUtil.SetRoleApis(ctx, role.Code, role.Apis); err != nil {
			slog.Error("缓存角色API权限失败", "roleCode", role.Code, "error", err)
		}
	}

	return nil
}

// ReloadUserRoles 重新加载用户的角色信息到缓存
// 当用户角色发生变化时调用此函数
func ReloadUserRoles(userID int) error {
	ctx := global.Context()

	// 从数据库中查询用户角色
	var user *po.UserModel
	var err error
	u := bo.UserModel
	user, err = u.Preload(u.Roles.RelationField).Where(u.ID.Eq(userID)).First()
	if err != nil {
		return err
	}

	// 提取角色编码
	roleCodes := lo.Map(user.Roles, func(role *po.RoleModel, _ int) string {
		return role.Code
	})

	// 更新缓存
	return redisUtil.SetUserRoles(ctx, userID, roleCodes)
}

// ReloadRoleApis 重新加载角色的API权限到缓存
func ReloadRoleApis(roleCode string) error {
	ctx := global.Context()

	// 用gorm gen链式API查询角色及其Apis
	role, err := bo.Q.RoleModel.Where(bo.Q.RoleModel.Code.Eq(roleCode)).Preload(bo.Q.RoleModel.Apis.RelationField).First()
	if err != nil {
		return err
	}

	// 更新缓存
	return redisUtil.SetRoleApis(ctx, roleCode, role.Apis)
}

// ReloadAllApiCache 重新加载所有API缓存
// 当API发生变化时调用此函数
func ReloadAllApiCache() error {
	ctx := global.Context()
	return loadAllApis(ctx)
}
