package rbac

import (
	"soybean-admin-go/model/dto"
	"soybean-admin-go/pkg/ginplus"
)

// OtherAPI 其它API控制器
type OtherAPI struct {
}

// CustomError 自定义错误
func (api *OtherAPI) CustomError(c *ginplus.Context) {
	var req dto.CustomErrorDTO
	if c.BindQuery(&req) {
		return
	}
	c.FailCode(req.Code, req.Msg)
}
