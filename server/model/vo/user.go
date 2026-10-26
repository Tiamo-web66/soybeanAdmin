package vo

// LoginVO 登录响应
type LoginVO struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// UserInfoVO 用户信息响应
type UserInfoVO struct {
	UserId   string   `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
	Buttons  []string `json:"buttons"`
}

type UserItemVO struct {
	Id         int      `json:"id"`
	CreateBy   string   `json:"createBy"` // = 联表查询的CreateBy user对象取出nickName
	CreateTime string   `json:"createTime"`
	UpdateBy   string   `json:"updateBy"`
	UpdateTime string   `json:"updateTime"`
	Status     string   `json:"status"`
	UserName   string   `json:"userName"`
	UserGender string   `json:"userGender"`
	NickName   string   `json:"nickName"`
	UserPhone  string   `json:"userPhone"`
	UserEmail  string   `json:"userEmail"`
	UserRoles  []string `json:"userRoles"`
}
