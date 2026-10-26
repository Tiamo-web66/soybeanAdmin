package dto

type CustomErrorDTO struct {
	Code int    `json:"code" form:"code"`
	Msg  string `json:"msg" form:"msg"`
}
