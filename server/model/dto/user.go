package dto

import (
	"soybean-admin-go/pkg/ginplus"
)

// LoginDTO 登录请求参数
type LoginDTO struct {
	UserName string `json:"userName" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// RefreshTokenReq 刷新令牌请求参数
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RegisterUserDTO 注册用户请求参数
type RegisterUserDTO struct {
	UserName string `json:"userName" binding:"required,min=3,max=20" example:"newuser"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// UpdatePasswordDTO 更新密码请求参数
type UpdatePasswordDTO struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"123456"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"newpass123"`
}

// GetUserListDTO 获取用户列表请求参数
// userName=null&userGender=null&nickName=null&userPhone=null&userEmail=null&userStatus=null
type GetUserListDTO struct {
	ginplus.PageDTO
	UserName   string  `json:"userName"  form:"userName"`
	UserGender *string `json:"userGender"  form:"userGender"`
	NickName   string  `json:"nickName"  form:"nickName"`
	UserPhone  string  `json:"userPhone"  form:"userPhone"`
	UserEmail  string  `json:"userEmail"  form:"userEmail"`
	UserStatus *string `json:"userStatus"  form:"userStatus"`
}

// CreateUserDTO 创建用户请求参数
type CreateUserDTO struct {
	UserName   string   `json:"userName" binding:"required,min=3,max=20" example:"newuser"`
	NickName   string   `json:"nickName"  example:"New User"`
	UserPhone  string   `json:"userPhone"  example:"13888888888"`
	UserEmail  string   `json:"userEmail"  example:"newuser@example.com"`
	UserGender string   `json:"userGender"  example:"1"`
	Status     string   `json:"status" binding:"required,oneof=0 1" example:"1"`
	UserRoles  []string `json:"userRoles" `
}

// UpdateUserDTO 更新用户请求参数
type UpdateUserDTO struct {
	CreateUserDTO
	Id int `json:"id" binding:"required"`
}
