package ginplus

import (
	"gorm.io/gen"
)

type IPage interface {
	Copy() *PageDTO //返回一个拷贝
	Constraint()    //对分页数据进行约束
}

// PageDTO 分页查询的请求数据
type PageDTO struct {
	Current int `json:"current" form:"current"`
	Size    int `json:"size" form:"size"`
}

func (d *PageDTO) Copy() *PageDTO {
	return d
}

func (d *PageDTO) Constraint() {
	// 对分页数据进行约束
	if d.Current < 1 {
		d.Current = 1
	}
	// 默认每页20条
	if d.Size <= 0 {
		d.Size = 20
	}
	// 最大每页100条
	if d.Size > 100 {
		d.Size = 100
	}
}

// Scope 返回 GORM 的分页范围函数
func (d *PageDTO) Scope() func(db gen.Dao) gen.Dao {
	return func(db gen.Dao) gen.Dao {
		// 计算偏移量
		offset := (d.Current - 1) * d.Size
		if offset < 0 {
			offset = 0
		}
		return db.Offset(offset).Limit(d.Size)
	}
}

// PageVO 分页查询的响应数据
type PageVO[T any] struct {
	Current int   `json:"current"`
	Size    int   `json:"size"`
	Total   int64 `json:"total"`
	Records T     `json:"records"`
}

// Resp 通用响应结构
type Resp[T any] struct {
	Code string `json:"code"` // 0表示成功，非0表示失败
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
