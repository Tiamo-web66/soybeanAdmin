package po

import "github.com/samber/lo"

// RoleModel 角色模型
// 与用户模型是多对多的关系
// 与菜单模型是多对多的关系
// 与权限模型是一对多的关系
// 与按钮模型是多对多的关系
type RoleModel struct {
	Model
	RbacOperatorModel
	Name       string         `json:"name"       gorm:"size:50;not null;uniqueIndex;comment:角色名称"`
	Code       string         `json:"code"       gorm:"size:50;not null;uniqueIndex;comment:角色编码"`
	Desc       string         `json:"desc"       gorm:"size:100;comment:角色描述"`
	Status     string         `json:"status"     gorm:"default:1;comment:是否启用0禁用1启用"`
	Menus      []*MenuModel   `json:"menus"      gorm:"many2many:l_role_menus"`
	Apis       []*ApiModel    `json:"apis"       gorm:"many2many:l_role_apis"`
	Buttons    []*ButtonModel `json:"buttons"    gorm:"many2many:l_role_buttons"`
	HomeMenuId int            `json:"homeMenuId" gorm:"default:1;comment:首页菜单id"`
	HomeMenu   *MenuModel     `json:"homeMenu"   gorm:"foreignKey:HomeMenuId;references:ID"`
	Users      []*UserModel   `json:"-" gorm:"many2many:l_user_roles;"`
}

// MenuModel 菜单模型(前端路由)
// 和角色模型是多对多的关系
type MenuModel struct {
	Model
	RbacOperatorModel
	ParentId        int             `json:"parentId" gorm:"not null;comment:父菜单ID,如果为0则为一级菜单"`
	MenuType        string          `json:"menuType" gorm:"not null;comment:菜单类型1目录2菜单"`
	MenuName        string          `json:"menuName" gorm:"not null;comment:菜单名称"`
	RouteName       string          `json:"routeName" gorm:"not null;uniqueIndex;comment:路由名称"`
	RoutePath       string          `json:"routePath" gorm:"not null;comment:路由路径"`
	Component       string          `json:"component" gorm:"not null;comment:组件名称"` //组件名称 layout.base$view.home 带$是在home组件外层包裹一层叫base的layout
	I18NKey         string          `json:"i18nKey" gorm:"not null;comment:国际化key"`
	IconType        string          `json:"iconType" gorm:"not null;comment:图标类型1iconify图标你2本地图标"`
	Icon            string          `json:"icon" gorm:"size:50;not null;comment:图标"`
	Status          string          `json:"status" gorm:"size:1;not null;default:1;comment:菜单状态1启用2禁用"`
	KeepAlive       bool            `json:"keepAlive" gorm:"not null;comment:是否缓存路由"`
	Constant        bool            `json:"constant" gorm:"not null;comment:是否常量路由"` //常量路由在应用启动时就被加载，适合用于不经常变动的基础页面。
	OrderBy         int             `json:"order" gorm:"not null;comment:排序,在同级路由中，越小越靠前"`
	Href            string          `json:"href" gorm:"size:100;not null;comment:外链地址"`
	FixedIndexInTab int             `json:"fixedIndexInTab" gorm:"not null;comment:固定在标签页中的序号"`
	HideInMenu      bool            `json:"hideInMenu" gorm:"not null;comment:是否隐藏菜单"`
	ActiveMenu      string          `json:"activeMenu" gorm:"not null;comment:高亮的菜单"` //路由页显示的时候,左侧的哪个菜单高亮,只有《隐藏菜单》为true时,这个才生效
	MultiTab        bool            `json:"multiTab" gorm:"not null;comment:是否支持多页签"`
	Query           RouteParamJsonB `json:"query" gorm:"type:jsonb;comment:路由参数"`
	//Buttons         []ButtonDTO     `json:"buttons"`         //按钮
	Buttons  []*ButtonModel `json:"buttons" gorm:"foreignKey:MenuId"`
	Apis     []*ApiModel    `json:"apis" gorm:"foreignKey:MenuId"`
	Children []*MenuModel   `json:"children" gorm:"foreignKey:ParentId"`
}

// ApiModel 权限模型(后端路由)
// 与菜单模型是多对一对关系
// 与角色模型是多对多的关系
type ApiModel struct {
	Model
	MenuId int    `json:"menuId"  gorm:"comment:菜单id"`
	Name   string `json:"name"    gorm:"size:50;not null;comment:权限名称"`
	Desc   string `json:"desc"    gorm:"size:100;comment:权限描述"`
	Method string `json:"method"  gorm:"uniqueIndex:key_method_path;size:50;not null;comment:HTTP方法(GET,POST,PUT,DELETE)"`
	Path   string `json:"path"    gorm:"uniqueIndex:key_method_path;size:100;not null;comment:路由路径(比如/api/users)"`
	Auth   string `json:"auth"    gorm:"size:1;default:2;comment:是否需要权限1需要2不需要"`
}

// ButtonModel 按钮模型,角色所拥有的按钮,在添加菜单那里添加,无法修改,可以被权限绑定和接触,但是不可以修改
// 与菜单模型是多对一对关系
// 与角色模型是多对多的关系
type ButtonModel struct {
	Model
	MenuId int    `json:"menuId"    gorm:"comment:菜单id"`
	Label  string `json:"label"     gorm:"size:50;not null;comment:按钮名"`
	Code   string `json:"code"      gorm:"size:50;not null;uniqueIndex;comment:按钮代码"`
}

func (RoleModel) TableName() string {
	return "l_role"
}

func (MenuModel) TableName() string {
	return "l_menu"
}

func (ApiModel) TableName() string { return "l_api" }

func (ButtonModel) TableName() string { return "l_button" }

const (
	SuperRoleCode     = "R_SUPER" // 超级管理员角色
	UserGenderMan     = "1"       // 男
	UserGenderWoman   = "2"       // 女
	UserStatusEnable  = "1"       // 用户启用
	MenuStatusEnable  = "1"       //菜单启用
	RoleStatusEnable  = "1"       // 角色启用
	ApiRequireAuth    = "1"       // 需要权限
	ApiNotRequireAuth = "2"       // 不需要权限
)

// RbacOperatorModel 有需要记录操作人的模型(为什么不和user表中一样,因为容易导致gorm的belongs to和has one不明确)
type RbacOperatorModel struct {
	CreateByUserId int        `json:"createByUserId" gorm:"comment:创建人id"`
	CreatedBy      *UserModel `json:"createdBy"  gorm:"foreignKey:CreateByUserId;references:ID"` //belongs to

	UpdateByUserId int        `json:"updateByUserId" gorm:"comment:更新人id"`
	UpdatedBy      *UserModel `json:"updatedBy"  gorm:"foreignKey:UpdateByUserId;references:ID"` //belongs to
}

// SafeCreateByNickName 安全的返回创建人昵称
func (r *RbacOperatorModel) SafeCreateByNickName() string {
	return lo.IfF(r.CreatedBy != nil, func() string { return r.CreatedBy.NickName }).Else("")
}

// SafeUpdateByNickName  安全的返回更新人昵称
func (r *RbacOperatorModel) SafeUpdateByNickName() string {
	return lo.IfF(r.UpdatedBy != nil, func() string { return r.UpdatedBy.NickName }).Else("")
}
