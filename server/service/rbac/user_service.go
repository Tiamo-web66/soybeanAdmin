package rbac

import (
	"errors"
	"log/slog"
	"soybean-admin-go/global"
	"soybean-admin-go/model/bo"
	"soybean-admin-go/model/dto"
	"soybean-admin-go/model/po"
	"soybean-admin-go/model/vo"
	"soybean-admin-go/utils/jwt"
	"soybean-admin-go/utils/password"
	redisUtil "soybean-admin-go/utils/redis"
	"strconv"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

// UserService 用户相关服务
type UserService struct{}

// Login 处理用户登录
func (s *UserService) Login(req *dto.LoginDTO) (*vo.LoginVO, error) {
	// 查询用户
	user, err := bo.UserModel.
		Preload(bo.UserModel.Roles).
		Where(bo.UserModel.UserName.Eq(req.UserName)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 验证密码
	if err := password.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, errors.New("密码错误")
	}

	// 更新最后登录时间
	_, err = bo.UserModel.Where(bo.UserModel.ID.Eq(user.ID)).UpdateSimple(
		bo.UserModel.LastLoginTime.Value(time.Now().Unix()),
	)
	if err != nil {
		slog.Error("更新最后登录时间失败", "error", err, "userId", user.ID)
	}

	// 提取角色代码
	roles := lo.Map(user.Roles, func(role *po.RoleModel, _ int) string {
		return role.Code
	})

	// 生成Token
	accessToken, err := jwt.GenerateToken(uint(user.ID), user.UserName, roles, jwt.AccessToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateToken(uint(user.ID), user.UserName, roles, jwt.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 将用户角色保存到Redis缓存中
	ctx := global.Context()
	if err := redisUtil.SetUserRoles(ctx, user.ID, roles); err != nil {
		// 记录错误但不影响登录流程
		slog.Error("缓存用户角色失败", "error", err, "userId", user.ID)
	}

	return &vo.LoginVO{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID int) (*vo.UserInfoVO, error) {
	user, err := bo.UserModel.
		Preload(bo.UserModel.Roles).
		Where(bo.UserModel.ID.Eq(userID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 获取角色编码
	roles := lo.Map(user.Roles, func(role *po.RoleModel, _ int) string {
		return role.Code
	})

	// 获取按钮权限
	var buttons []string
	for _, role := range user.Roles {
		// 预加载每个角色的Buttons
		roleWithButtons, err := bo.RoleModel.
			Preload(bo.RoleModel.Buttons).
			Where(bo.RoleModel.ID.Eq(role.ID)).
			First()
		if err == nil && roleWithButtons != nil {
			for _, button := range roleWithButtons.Buttons {
				buttons = append(buttons, button.Code)
			}
		}
	}
	// 去重
	buttons = lo.Uniq(buttons)

	return &vo.UserInfoVO{
		UserId:   strconv.Itoa(user.ID),
		UserName: user.UserName,
		Roles:    roles,
		Buttons:  buttons,
	}, nil
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(refreshToken string) (*vo.LoginVO, error) {
	// 解析并验证refreshToken
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 验证token类型
	if claims.RegisteredClaims.Subject != string(jwt.RefreshToken) {
		return nil, errors.New("无效的刷新令牌")
	}

	// 从数据库获取最新的用户信息和角色
	user, err := bo.UserModel.
		Preload(bo.UserModel.Roles).
		Where(bo.UserModel.ID.Eq(int(claims.UserID))).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 提取角色代码
	roles := lo.Map(user.Roles, func(role *po.RoleModel, _ int) string {
		return role.Code
	})

	// 生成新的token
	newAccessToken, err := jwt.GenerateToken(uint(user.ID), user.UserName, roles, jwt.AccessToken)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := jwt.GenerateToken(uint(user.ID), user.UserName, roles, jwt.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 更新Redis中的用户角色缓存
	ctx := global.Context()
	if err := redisUtil.SetUserRoles(ctx, user.ID, roles); err != nil {
		// 记录错误但不影响刷新令牌流程
		slog.Error("刷新令牌时更新用户角色缓存失败", "error", err, "userId", user.ID)
	}

	return &vo.LoginVO{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// UpdateUserPassword 更新用户密码
func (s *UserService) UpdateUserPassword(userID int, req *dto.UpdatePasswordDTO) error {
	// 查询用户
	user, err := bo.UserModel.
		Where(bo.UserModel.ID.Eq(userID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if err := password.VerifyPassword(user.Password, req.OldPassword); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	_, err = bo.UserModel.
		Where(bo.UserModel.ID.Eq(userID)).
		UpdateSimple(
			bo.UserModel.Password.Value(hashedPassword),
		)
	return err
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(req *dto.CreateUserDTO, operatorID int) (int, error) {
	// 检查用户名是否已存在
	count, err := bo.UserModel.
		Where(bo.UserModel.UserName.Eq(req.UserName)).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("用户名已存在")
	}

	// 检查是否尝试创建超级管理员
	for _, roleCode := range req.UserRoles {
		if roleCode == po.SuperRoleCode {
			return 0, errors.New("不允许创建超级管理员角色的用户")
		}
	}

	// 加密默认密码 123456
	hashedPassword, err := password.HashPassword("123456")
	if err != nil {
		return 0, err
	}

	// 创建用户对象
	user := &po.UserModel{
		UserName:   req.UserName,
		Password:   hashedPassword,
		NickName:   req.NickName,
		UserPhone:  req.UserPhone,
		UserEmail:  req.UserEmail,
		UserGender: req.UserGender,
		Status:     req.Status,
		OperatorModel: po.OperatorModel{
			CreateById: operatorID,
		},
	}

	// 开始事务
	err = bo.Q.Transaction(func(tx *bo.Query) error {
		// 创建用户
		if err := tx.UserModel.Create(user); err != nil {
			return err
		}

		// 如果指定了角色代码，查询并关联角色
		if len(req.UserRoles) > 0 {
			roles, err := tx.RoleModel.
				Where(tx.RoleModel.Code.In(req.UserRoles...)).
				Find()
			if err != nil {
				return err
			}

			// 关联用户和角色
			if len(roles) > 0 {
				err = tx.UserModel.Roles.Model(user).Append(roles...)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(req *dto.UpdateUserDTO, operatorID int) error {
	// 事务处理
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询用户是否存在，同时预加载角色
		user, err := tx.UserModel.
			Preload(tx.UserModel.Roles).
			Where(tx.UserModel.ID.Eq(req.Id)).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		// 检查是否尝试修改超级管理员
		var isCurrentUserSuperAdmin bool

		// 检查当前用户是否为超级管理员
		for _, role := range user.Roles {
			if role.Code == po.SuperRoleCode {
				isCurrentUserSuperAdmin = true
				break
			}
		}

		// 检查是否在请求中包含超级管理员角色
		var requestHasSuperAdmin bool
		for _, roleCode := range req.UserRoles {
			if roleCode == po.SuperRoleCode {
				requestHasSuperAdmin = true
				break
			}
		}

		// 验证修改是否符合规则
		if isCurrentUserSuperAdmin && !requestHasSuperAdmin {
			return errors.New("不能将超级管理员降级为普通用户")
		}

		if !isCurrentUserSuperAdmin && requestHasSuperAdmin {
			return errors.New("不能将普通用户提升为超级管理员")
		}

		// 检查用户名是否重复（如果修改了用户名）
		if req.UserName != user.UserName {
			count, err := tx.UserModel.Where(tx.UserModel.UserName.Eq(req.UserName), tx.UserModel.ID.Neq(req.Id)).Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return errors.New("用户名已存在")
			}
		}

		// 更新用户基本信息
		_, err = tx.UserModel.Where(tx.UserModel.ID.Eq(req.Id)).
			UpdateSimple(
				tx.UserModel.UserName.Value(req.UserName),
				tx.UserModel.NickName.Value(req.NickName),
				tx.UserModel.UserPhone.Value(req.UserPhone),
				tx.UserModel.UserEmail.Value(req.UserEmail),
				tx.UserModel.UserGender.Value(req.UserGender),
				tx.UserModel.Status.Value(req.Status),
				tx.UserModel.UpdateById.Value(operatorID),
			)
		if err != nil {
			return err
		}

		// 获取当前用户的角色代码列表，用于后续比较
		currentRoleCodes := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			currentRoleCodes = append(currentRoleCodes, role.Code)
		}

		// 检查角色是否有变动
		roleChanged := false

		// 如果长度不同，肯定有变化
		if len(currentRoleCodes) != len(req.UserRoles) {
			roleChanged = true
		} else {
			// 比较两个切片的内容
			roleCodeSet := make(map[string]struct{}, len(currentRoleCodes))
			for _, code := range currentRoleCodes {
				roleCodeSet[code] = struct{}{}
			}

			// 检查请求的角色是否都在当前角色集合中
			for _, reqCode := range req.UserRoles {
				if _, exists := roleCodeSet[reqCode]; !exists {
					roleChanged = true
					break
				}
			}
		}

		// 更新角色关联
		if len(req.UserRoles) > 0 {
			// 查询角色
			roles, err := tx.RoleModel.Where(tx.RoleModel.Code.In(req.UserRoles...)).Find()
			if err != nil {
				return err
			}

			// 使用Replace直接替换角色关联，避免使用Clear+Append组合
			if len(roles) > 0 {
				err = tx.UserModel.Roles.Model(user).Replace(roles...)
				if err != nil {
					return err
				}
			}
		} else {
			// 如果角色列表为空，则清空所有角色关联
			err = tx.UserModel.Roles.Model(user).Clear()
			if err != nil {
				return err
			}
			// 如果请求的角色为空，而之前有角色，则认为角色发生了变化
			roleChanged = len(currentRoleCodes) > 0
		}

		// 如果角色有变动，同步更新Redis缓存
		if roleChanged {
			ctx := global.Context()

			// 如果是清空角色，则传入空数组
			updatedRoleCodes := req.UserRoles
			if len(updatedRoleCodes) == 0 {
				updatedRoleCodes = []string{}
			}

			// 更新Redis缓存
			if err := redisUtil.SetUserRoles(ctx, user.ID, updatedRoleCodes); err != nil {
				// 记录错误但不影响事务提交
				slog.Error("更新用户角色缓存失败", "error", err, "userId", user.ID)
			}
		}

		return nil
	})
}

// DeleteUser 删除单个用户
func (s *UserService) DeleteUser(userId int) error {
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询用户是否存在，同时预加载角色
		user, err := tx.UserModel.
			Preload(tx.UserModel.Roles).
			Where(tx.UserModel.ID.Eq(userId)).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		// 检查是否尝试删除超级管理员
		if po.IsSuper(user.Roles...) {
			return errors.New("不能删除超级管理员账户")
		}

		// 清除用户的角色关联
		err = tx.UserModel.Roles.Model(user).Clear()
		if err != nil {
			return err
		}

		// 删除用户
		_, err = tx.UserModel.Where(tx.UserModel.ID.Eq(userId)).Delete()
		return err
	})
}

// DeleteUsers 批量删除用户
func (s *UserService) DeleteUsers(userIds []int) error {
	if len(userIds) == 0 {
		return errors.New("用户ID列表不能为空")
	}
	return bo.Q.Transaction(func(tx *bo.Query) error {
		// 查询用户是否存在，同时预加载角色
		users, err := tx.UserModel.Preload(tx.UserModel.Roles).Where(tx.UserModel.ID.In(userIds...)).Find()
		if err != nil {
			return err
		}
		if len(users) != len(userIds) {
			return errors.New("部分用户不存在")
		}

		// 检查是否尝试删除超级管理员
		for _, user := range users {
			if po.IsSuper(user.Roles...) {
				return errors.New("不能删除超级管理员账户")
			}
		}

		// 清除所有用户的角色关联 - 优化为单次操作
		if len(users) > 0 {
			// 使用UnderlyingDB来执行一条批量删除语句，避免循环调用Clear()
			query := tx.UserModel.UnderlyingDB().
				Table("l_user_roles").
				Where("user_model_id IN ?", userIds)

			// 执行批量删除 - 这种方式更高效
			err = query.Unscoped().Delete("").Error
			if err != nil {
				return err
			}
		}

		// 批量删除用户
		_, err = tx.UserModel.Where(tx.UserModel.ID.In(userIds...)).Delete()
		return err
	})
}
