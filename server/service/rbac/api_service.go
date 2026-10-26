package rbac

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log/slog"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/po"
	"soybean-admin-go/pkg/ginplus"
	"time"

	"github.com/samber/lo"

	"gorm.io/gorm/clause"
)

// APIService 处理API相关的业务逻辑
type APIService struct{}

// SyncAPIToDatabase 将收集到的API信息同步到数据库
func (s *APIService) SyncAPIToDatabase(engine *ginplus.Engine, menuID int) error {
	apis := engine.GetAPIs()
	if len(apis) == 0 {
		return nil
	}

	// 提取所有当前有效API的路径和方法
	var allApiPaths []string
	for _, api := range apis {
		allApiPaths = append(allApiPaths, api.Method+":"+api.Path)
	}

	// 获取所有数据库中已存在的API
	q := bo.Q
	existingAPIs, err := q.ApiModel.Where(q.ApiModel.MenuId.Eq(menuID)).Find()
	if err != nil {
		return fmt.Errorf("获取现有API失败: %w", err)
	}

	// 构建路径映射
	pathToAPI := make(map[string]*po.ApiModel)
	for _, api := range existingAPIs {
		key := api.Method + ":" + api.Path
		pathToAPI[key] = api
	}

	// 准备要保留的API模型
	var validApiModels []*po.ApiModel
	nowTime := time.Now().UnixMilli()

	for _, api := range apis {
		key := api.Method + ":" + api.Path
		if existing, ok := pathToAPI[key]; ok {
			// 已存在的API，保留ID
			validApiModels = append(validApiModels, &po.ApiModel{
				Model: po.Model{
					ID:        existing.ID,
					CreatedAt: existing.CreatedAt,
					UpdatedAt: nowTime,
				},
				MenuId: menuID,
				Name:   api.Name,
				Desc:   api.Desc,
				Method: api.Method,
				Path:   api.Path,
				Auth:   lo.Ternary(api.RequireAuth, po.ApiRequireAuth, po.ApiNotRequireAuth),
			})
		} else {
			// 新API
			validApiModels = append(validApiModels, &po.ApiModel{
				Model: po.Model{
					CreatedAt: nowTime,
					UpdatedAt: nowTime,
				},
				MenuId: menuID,
				Name:   api.Name,
				Desc:   api.Desc,
				Method: api.Method,
				Path:   api.Path,
				Auth:   lo.Ternary(api.RequireAuth, po.ApiRequireAuth, po.ApiNotRequireAuth),
			})
		}
	}

	err = s.InstallOrUpdateAPIs(apis, menuID)
	if err != nil {
		return err
	}
	return nil
}

// SyncAuthAPIToDatabase 将需要认证的API信息同步到数据库
func (s *APIService) SyncAuthAPIToDatabase(engine *ginplus.Engine, menuID int) error {
	apis := engine.GetAuthAPIs()
	if len(apis) == 0 {
		return nil
	}

	return s.InstallOrUpdateAPIs(apis, menuID)
}

// SyncPublicAPIToDatabase 将不需要认证的API信息同步到数据库
func (s *APIService) SyncPublicAPIToDatabase(engine *ginplus.Engine, menuID int) error {
	apis := engine.GetPublicAPIs()
	if len(apis) == 0 {
		return nil
	}

	return s.InstallOrUpdateAPIs(apis, menuID)
}

// InstallOrUpdateAPIs 将API信息安装或更新到数据库
func (s *APIService) InstallOrUpdateAPIs(apis []ginplus.APIInfo, menuID int) error {
	if len(apis) == 0 {
		return nil
	}

	q := bo.Q
	if q == nil {
		return fmt.Errorf("数据库查询对象为空")
	}

	// 当前时间戳
	nowTime := time.Now().UnixMilli()

	// 准备API记录
	var apiModels []*po.ApiModel

	for _, api := range apis {
		apiModel := &po.ApiModel{
			Model: po.Model{
				CreatedAt: nowTime,
				UpdatedAt: nowTime,
			},
			MenuId: menuID,
			Name:   api.Name,
			Desc:   api.Desc,
			Method: api.Method,
			Path:   api.Path,
			Auth:   lo.Ternary(api.RequireAuth, po.ApiRequireAuth, po.ApiNotRequireAuth),
		}
		apiModels = append(apiModels, apiModel)
	}

	// 使用事务进行操作
	return q.Transaction(func(tx *bo.Query) error {
		// 使用 Clauses 和 OnConflict 进行 UPSERT 操作
		// 冲突时更新除 menuId 外的所有字段
		if err := tx.ApiModel.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "method"}, {Name: "path"}},                        // 冲突判断字段
			DoUpdates: clause.AssignmentColumns([]string{"name", "desc", "auth", "updated_at"}), // 仅更新这些字段，不包含 menu_id
		}).Create(apiModels...); err != nil {
			return fmt.Errorf("UPSERT API记录失败: %w", err)
		}

		// 清理过时的API
		return s.CleanObsoleteAPIs(apiModels)
	})
}

// CleanObsoleteAPIs 清理不再使用的API
// menuID: 菜单ID
// apiModels: 有效的API模型列表
func (s *APIService) CleanObsoleteAPIs(apiModels []*po.ApiModel) error {
	q := bo.Q

	// 构建有效API的ID映射
	validIDs := make(map[int]bool)
	for _, api := range apiModels {
		if api.ID > 0 { // 只有ID大于0的才是已存在于数据库的记录
			validIDs[api.ID] = true
		}
	}

	// 查询所有API
	APIs, err := q.ApiModel.Find()
	if err != nil {
		return fmt.Errorf("查询所有API失败: %w", err)
	}

	// 找出需要删除的API（即不在有效ID列表中的API）
	var obsoleteAPIs []*po.ApiModel
	for _, api := range APIs {
		if !validIDs[api.ID] {
			obsoleteAPIs = append(obsoleteAPIs, api)
		}
	}

	// 如果没有需要删除的API，直接返回
	if len(obsoleteAPIs) == 0 {
		return nil
	}

	// 删除过时的API（使用事务外的查询对象避免死锁）
	err = q.ApiModel.UnderlyingDB().
		Where("id IN ?", lo.Map(obsoleteAPIs, func(api *po.ApiModel, _ int) int { return api.ID })).
		Delete(&po.ApiModel{}).Error
	if err != nil {
		return fmt.Errorf("删除过时API失败: %w", err)
	}

	// 删除角色与API的关联关系
	apiIDs := lo.Map(obsoleteAPIs, func(api *po.ApiModel, _ int) int { return api.ID })
	if len(apiIDs) > 0 {
		err = q.ApiModel.UnderlyingDB().
			Table("l_role_apis").
			Where("api_model_id IN ?", apiIDs).
			Unscoped().
			Delete("").Error
		if err != nil {
			return fmt.Errorf("删除角色与API的关联关系失败: %w", err)
		}
		fmt.Printf("成功删除 %d 个API相关的角色关联\n", len(apiIDs))
	}

	fmt.Printf("成功删除 %d 个过时API\n", len(obsoleteAPIs))

	// 如果删除了API，需要更新Redis缓存

	if err := s.refreshRbacCache(); err != nil {
		// 只记录错误，不影响正常流程
		fmt.Printf("更新RBAC缓存失败: %v\n", err)
	}

	return nil
}

// 更新RBAC缓存
func (s *APIService) refreshRbacCache() error {
	// 导入初始化包，调用重新加载API缓存的函数
	err := initialize.ReloadAllApiCache()
	if err != nil {
		return fmt.Errorf("重新加载API缓存失败: %w", err)
	}
	return nil
}

// NewAPIService 创建新的API服务实例
func NewAPIService() *APIService {
	return &APIService{}
}

// GetRoleApis 获取角色拥有的API权限
func (s *APIService) GetRoleApis(roleId int, auth string) ([]int, error) {
	r := bo.RoleModel
	a := bo.ApiModel
	query := r.Where(r.ID.Eq(roleId))
	if auth != "" {
		query = query.Preload(r.Apis.On(a.Auth.Eq(auth)))
	} else {
		query = query.Preload(r.Apis)
	}
	role, err := query.First()
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}
	//如果是超管 查询出所有的api
	if po.IsSuper(role) {
		var apiAll []*po.ApiModel
		if auth != "" {
			apiAll, err = bo.ApiModel.Where(bo.ApiModel.Auth.Eq(auth)).Find()
		} else {
			apiAll, err = bo.ApiModel.Find()
		}
		if err != nil {
			return nil, err
		}
		role.Apis = apiAll
	}
	ids := lo.Map(role.Apis, func(api *po.ApiModel, _ int) int {
		return api.ID
	})
	return ids, nil
}

// AssignRoleApis 为角色分配API权限
func (s *APIService) AssignRoleApis(roleID int, apiIDs []int) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		//先查出角色 和 角色关联的api
		role, err := tx.RoleModel.Where(tx.RoleModel.ID.Eq(roleID)).
			Preload(tx.RoleModel.Apis).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("角色不存在")
			}
			return err
		}
		if po.IsSuper(role) {
			return errors.New("不允许修改超级管理员角色")
		}

		//查询本次重新分配的api
		var apis []*po.ApiModel
		if len(apiIDs) > 0 {
			apis, err = tx.ApiModel.Where(tx.ApiModel.ID.In(apiIDs...)).Find()
			if err != nil {
				return err
			}
			if len(apis) != len(apiIDs) {
				return errors.New("部分API不存在")
			}
		}
		// 重新分配
		if err = tx.RoleModel.Apis.Model(role).Replace(apis...); err != nil {
			return err
		}
		// 更新角色API缓存
		if err = initialize.ReloadRoleApis(role.Code); err != nil {
			slog.Error("更新角色API缓存失败", "roleCode", role.Code, "error", err)
			return err
		}
		return nil
	})
}
