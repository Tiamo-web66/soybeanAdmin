package vo

type RoleItemVO struct {
	Id         int    `json:"id"`
	CreateBy   string `json:"createBy"`
	CreateTime string `json:"createTime"`
	UpdateBy   string `json:"updateBy"`
	UpdateTime string `json:"updateTime"`
	Status     string `json:"status"`
	RoleName   string `json:"roleName"`
	RoleCode   string `json:"roleCode"`
	RoleDesc   string `json:"roleDesc"`
	HomeMenuId int    `json:"homeMenuId"`
}

type RoleSimplifyItemVO struct {
	Id       string `json:"id"`
	RoleName string `json:"roleName"`
	RoleCode string `json:"roleCode"`
}

type RoleMenuIdsVO struct {
	MenuIds       []int  `json:"menuIds"`
	HomeRouteName string `json:"homeRouteName"`
}
