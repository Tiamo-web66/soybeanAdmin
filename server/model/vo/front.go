package vo

type FrontRouteMetaVO struct {
	Title      string `json:"title"`
	I18NKey    string `json:"i18nKey"`
	Icon       string `json:"icon,omitempty"`
	Order      int    `json:"order,omitempty"`
	HideInMenu bool   `json:"hideInMenu,omitempty"`
	MultiTab   bool   `json:"multiTab,omitempty"`
	ActiveMenu string `json:"activeMenu,omitempty"`
	KeepAlive  bool   `json:"keepAlive,omitempty"`
	Constant   bool   `json:"constant,omitempty"`
}

type FrontRouteItemVO struct {
	Name      string              `json:"name"`
	Path      string              `json:"path"`
	Component string              `json:"component"`
	Redirect  string              `json:"redirect,omitempty"` //如果Component为空,将重定向到该路由Children中path
	Props     bool                `json:"props,omitempty"`    //判断path路径中是否有:,有就是props
	Meta      FrontRouteMetaVO    `json:"meta"`
	Children  []*FrontRouteItemVO `json:"children,omitempty"` //获取用户路由的api需要这一项
}

// FrontRouteVO 前端路由
type FrontRouteVO struct {
	Routes []*FrontRouteItemVO `json:"routes"`
	Home   string              `json:"home"` //找到Routes中第一个没有Children的name
}
