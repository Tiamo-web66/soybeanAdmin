package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"soybean-admin-go/global"
	"soybean-admin-go/model/po"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// RoleApisCacheKey 角色拥有的API缓存键前缀
	RoleApisCacheKey = "soybean:rbac:role:%s:apis"

	// UserRolesCacheKey 用户拥有的角色缓存键
	UserRolesCacheKey = "soybean:rbac:user:%d:roles"

	// ApiAuthCacheKey API权限验证缓存键
	ApiAuthCacheKey = "soybean:rbac:api:%s:%s"

	// CacheExpiration 缓存过期时间(12小时)
	CacheExpiration = 12 * time.Hour
)

// ErrCacheMiss 缓存未命中错误
var ErrCacheMiss = errors.New("cache miss")

// SetRoleApis 缓存角色拥有的API权限
func SetRoleApis(ctx context.Context, roleCode string, apis []*po.ApiModel) error {
	// 序列化API列表
	apisJson, err := json.Marshal(apis)
	if err != nil {
		return fmt.Errorf("序列化API失败: %w", err)
	}

	// 设置缓存
	key := fmt.Sprintf(RoleApisCacheKey, roleCode)
	return global.REDIS.Set(ctx, key, apisJson, CacheExpiration).Err()
}

// GetRoleApis 获取角色拥有的API权限
func GetRoleApis(ctx context.Context, roleCode string) ([]*po.ApiModel, error) {
	key := fmt.Sprintf(RoleApisCacheKey, roleCode)

	// 获取缓存
	result, err := global.REDIS.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	// 反序列化API列表
	var apis []*po.ApiModel
	if err := json.Unmarshal([]byte(result), &apis); err != nil {
		return nil, fmt.Errorf("反序列化API失败: %w", err)
	}

	return apis, nil
}

// SetUserRoles 缓存用户拥有的角色
func SetUserRoles(ctx context.Context, userID int, roleCodes []string) error {
	// 序列化角色编码列表
	rolesJson, err := json.Marshal(roleCodes)
	if err != nil {
		return fmt.Errorf("序列化角色编码失败: %w", err)
	}

	// 设置缓存
	key := fmt.Sprintf(UserRolesCacheKey, userID)
	return global.REDIS.Set(ctx, key, rolesJson, CacheExpiration).Err()
}

// GetUserRoles 获取用户拥有的角色
func GetUserRoles(ctx context.Context, userID int) ([]string, error) {
	key := fmt.Sprintf(UserRolesCacheKey, userID)

	// 获取缓存
	result, err := global.REDIS.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	// 反序列化角色编码列表
	var roleCodes []string
	if err := json.Unmarshal([]byte(result), &roleCodes); err != nil {
		return nil, fmt.Errorf("反序列化角色编码失败: %w", err)
	}

	return roleCodes, nil
}

// VerifyApiAuth 验证API访问权限
// method: HTTP方法(GET,POST,PUT,DELETE)
// path: API路径(/api/users等)
// roleCodes: 用户拥有的角色编码列表
func VerifyApiAuth(ctx context.Context, method, path string, roleCodes []string) (bool, error) {
	// 检查是否超级管理员角色
	for _, roleCode := range roleCodes {
		if roleCode == po.SuperRoleCode {
			return true, nil // 超级管理员拥有所有权限
		}
	}

	// 从缓存中检查API权限
	key := fmt.Sprintf(ApiAuthCacheKey, method, path)
	authValue, err := global.REDIS.Get(ctx, key).Result()
	if err == nil {
		// 缓存命中，检查API是否不需要权限
		if authValue == po.ApiNotRequireAuth {
			return true, nil // 不需要权限，允许访问
		}
	}

	// 检查角色是否有API访问权限
	for _, roleCode := range roleCodes {
		apis, err := GetRoleApis(ctx, roleCode)
		if err != nil {
			if errors.Is(err, ErrCacheMiss) {
				continue // 缓存未命中，检查下一个角色
			}
			return false, err
		}

		// 查找匹配的API
		for _, api := range apis {
			if api.Method == method && api.Path == path {
				return true, nil // 找到匹配的API，允许访问
			}
		}
	}

	return false, nil // 未找到匹配的API权限，拒绝访问
}

// PreCacheApiAuth 预先缓存API权限
func PreCacheApiAuth(ctx context.Context, api *po.ApiModel) error {
	key := fmt.Sprintf(ApiAuthCacheKey, api.Method, api.Path)
	return global.REDIS.Set(ctx, key, api.Auth, CacheExpiration).Err()
}

// ClearRoleApiCache 清除角色API缓存
func ClearRoleApiCache(ctx context.Context, roleCode string) error {
	key := fmt.Sprintf(RoleApisCacheKey, roleCode)
	return global.REDIS.Del(ctx, key).Err()
}

// ClearUserRoleCache 清除用户角色缓存
func ClearUserRoleCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(UserRolesCacheKey, userID)
	return global.REDIS.Del(ctx, key).Err()
}

// ClearApiAuthCache 清除API权限缓存
func ClearApiAuthCache(ctx context.Context, method, path string) error {
	key := fmt.Sprintf(ApiAuthCacheKey, method, path)
	return global.REDIS.Del(ctx, key).Err()
}
