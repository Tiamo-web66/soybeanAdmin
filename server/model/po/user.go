package po

import "github.com/samber/lo"

// UserModel 用户模型
type UserModel struct {
	Model
	OperatorModel
	UserName      string       `json:"userName" gorm:"size:50;not null;uniqueIndex"`
	Password      string       `json:"-" gorm:"size:100;not null"` // 密码不返回给前端
	Status        string       `json:"status"    gorm:"size:1;default:1;comment:是否启用0禁用1启用"`
	UserGender    string       `json:"userGender" gorm:"size:1;default:1;comment:性别1男2女"`
	NickName      string       `json:"nickName" gorm:"size:50;comment:昵称"`
	UserPhone     string       `json:"userPhone" gorm:"size:11;comment:手机号"`
	UserEmail     string       `json:"userEmail" gorm:"size:50;comment:邮箱"`
	LastLoginTime int64        `json:"lastLoginTime"`
	Roles         []*RoleModel `json:"roles" gorm:"many2many:l_user_roles;"`
}

// TableName 设置User表名
func (UserModel) TableName() string {
	return "l_users"
}

// OperatorModel 有需要记录操作人的模型
type OperatorModel struct {
	CreateById int        `json:"createById" gorm:"comment:创建人id"`
	CreatedBy  *UserModel `json:"createdBy"  gorm:"foreignKey:CreateById;references:ID"` //belongs to

	UpdateById int        `json:"updateById" gorm:"comment:更新人id"`
	UpdatedBy  *UserModel `json:"updatedBy"  gorm:"foreignKey:UpdateById;references:ID"` //belongs to
}

// SafeCreateByNickName 安全的返回创建人昵称
func (r *OperatorModel) SafeCreateByNickName() string {
	return lo.IfF(r.CreatedBy != nil, func() string { return r.CreatedBy.NickName }).Else("")
}

// SafeUpdateByNickName  安全的返回更新人昵称
func (r *OperatorModel) SafeUpdateByNickName() string {
	return lo.IfF(r.UpdatedBy != nil, func() string { return r.UpdatedBy.NickName }).Else("")
}
