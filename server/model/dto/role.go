package dto

import (
	"soybean-admin-go/model/vo"
	"soybean-admin-go/pkg/ginplus"
)

type GetRoleListDTO struct {
	ginplus.PageDTO
	RoleName string  `json:"roleName" form:"roleName"`
	RoleCode string  `json:"roleCode" form:"roleCode"`
	Status   *string `json:"status"  form:"status"`
}

type UpdateMenuIdByRoleDTO struct {
	vo.RoleMenuIdsVO
}
