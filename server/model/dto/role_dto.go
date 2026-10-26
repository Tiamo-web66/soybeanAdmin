package dto

// CreateRoleDTO 创建角色请求
type CreateRoleDTO struct {
	RoleName string `json:"roleName" binding:"required" form:"roleName"`
	RoleCode string `json:"roleCode" binding:"required" form:"roleCode"`
	Status   string `json:"status" binding:"required" form:"status"`
	RoleDesc string `json:"roleDesc" form:"roleDesc"`
}

// UpdateRoleDTO 更新角色请求
type UpdateRoleDTO struct {
	CreateRoleDTO
	Id         int  `json:"id" binding:"required"`
	HomeMenuId *int `json:"homeMenuId" `
}
