package dto

import (
	"soybean-admin-go/model/po"
	"soybean-admin-go/pkg/ginplus"
)

// BaseMenuDTO 基础菜单DTO
type BaseMenuDTO struct {
	Id              int                `json:"id"`              //菜单id
	ParentId        int                `json:"parentId"`        //父菜单id ,如果为0则为一级菜单
	MenuType        string             `json:"menuType"`        //菜单类型1：目录2：菜单
	MenuName        string             `json:"menuName"`        //菜单名称
	RouteName       string             `json:"routeName"`       //路由名称 home
	RoutePath       string             `json:"routePath"`       //路由路径 /home
	Component       string             `json:"component"`       //组件名称 layout.base$view.home 带$是在home组件外层包裹一层叫base的layout
	I18NKey         string             `json:"i18nKey"`         //国际化key
	IconType        string             `json:"iconType"`        //图标类型 1 iconify图标你 2 本地图标
	Icon            string             `json:"icon"`            //图标
	Status          string             `json:"status"`          //菜单状态1启用2禁用
	KeepAlive       bool               `json:"keepAlive"`       //是否缓存路由
	Constant        bool               `json:"constant"`        //是否常量路由,常量路由在应用启动时就被加载，适合用于不经常变动的基础页面。
	Order           int                `json:"order"`           //排序,在同级路由中，越小越靠前
	Href            string             `json:"href"`            //外链地址
	FixedIndexInTab int                `json:"fixedIndexInTab"` //固定在标签页中的序号
	HideInMenu      bool               `json:"hideInMenu"`      //是否隐藏菜单
	ActiveMenu      string             `json:"activeMenu"`      //高亮的菜单,路由页显示的时候,左侧的哪个菜单高亮,只有《隐藏菜单》为true时,这个才生效
	MultiTab        bool               `json:"multiTab"`        //是否支持多页签
	Query           po.RouteParamJsonB `json:"query"`           //路由参数
	Buttons         []ButtonDTO        `json:"buttons"`         //按钮
	//UpdateBy string `json:"updateBy"` //谁修改的,注意,这个不需要前端传递
	//Index    int    `json:"index"`
}

// ConvertRouteQueriesToRouteParamJsonB 将RouteQueryDTO.Query 转换为 RouteParamJsonB
//func (dto *BaseMenuDTO) ConvertRouteQueriesToRouteParamJsonB() po.RouteParamJsonB {
//	dst := lo.Map(dto.Query, func(item RouteQueryDTO, index int) *po.RouteParam {
//		return &po.RouteParam{
//			Key: item.Key,
//			Val: item.Value,
//		}
//	})
//	return dst
//}

// EditMenuDTO 修改菜单
type EditMenuDTO struct {
	BaseMenuDTO     // 嵌入基础菜单DTO
	Id          int `json:"id"` // 菜单id，仅EditMenuDTO需要
}

// AddMenuDTO 添加菜单
type AddMenuDTO struct {
	BaseMenuDTO // 嵌入基础菜单DTO
}

// GetMenuListDTO 查询菜单
type GetMenuListDTO struct {
	ginplus.PageDTO
}

// ButtonDTO 菜单所拥有的按钮
type ButtonDTO struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
}
