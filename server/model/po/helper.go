package po

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IsSuper 是否是超级管理员
func IsSuper(roles ...*RoleModel) bool {
	for _, role := range roles {
		if isSuper(role.Code) {
			return true
		}
	}
	return false
}

// IsSuperByCode 通过code判断是否是超级管理员
func IsSuperByCode(roleCodes ...string) bool {
	for _, code := range roleCodes {
		if isSuper(code) {
			return true
		}
	}
	return false
}

// isSuper 是否是超级管理员
func isSuper(roleCode string) bool {
	if roleCode == SuperRoleCode {
		return true
	}
	return false
}

type RouteParam struct {
	Key string `json:"key"`   //路由参数key
	Val string `json:"value"` //路由参数value
}

type RouteParamJsonB []*RouteParam

// Value 实现 driver.Valuer 接口，Value 返回 json value
//
//goland:noinspection GoMixedReceiverTypes 如果这里的类型为*RouteParamJsonB，那么就不会被gorm调用
func (r RouteParamJsonB) Value() (driver.Value, error) {
	if r == nil || len(r) == 0 {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
//
//goland:noinspection GoMixedReceiverTypes 这里必须为*RouteParamJsonB，因为需要把值赋给RouteParamJsonB
func (r *RouteParamJsonB) Scan(src any) error {
	if src == nil {
		*r = make([]*RouteParam, 0)
		return nil
	}

	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("不支持的类型: %T", src)
	}

	// 为空字符串或空JSON数组时，返回空切片
	if len(data) == 0 || string(data) == "" || string(data) == "[]" {
		*r = make([]*RouteParam, 0)
		return nil
	}

	return json.Unmarshal(data, r)
}
