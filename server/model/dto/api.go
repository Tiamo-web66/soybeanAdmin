package dto

import "soybean-admin-go/pkg/ginplus"

type GetApiListDTO struct {
	ginplus.PageDTO
	Name   *string `json:"name"`
	Method *string `json:"method"`
	Auth   *string `json:"auth"`
}
