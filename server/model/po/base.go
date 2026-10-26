package po

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

// Model 通用字段
type Model struct {
	ID        int                   `json:"id"        gorm:"primaryKey"`
	CreatedAt int64                 `json:"createdAt" gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt int64                 `json:"updatedAt" gorm:"autoUpdateTime:milli;comment:更新时间"`
	DeletedAt soft_delete.DeletedAt `json:"-"         gorm:"default:0;index;softDelete:milli;comment:是否删除"`
}

// CreateTime 创建时间 将毫秒时间戳转换为 YYYY-MM-DD HH:mm:ss
func (m *Model) CreateTime() string {
	if m.CreatedAt == 0 {
		return ""
	}
	return time.UnixMilli(m.CreatedAt).Format(time.DateTime)
}

// UpdateTime 更新时间 将毫秒时间戳转换为 YYYY-MM-DD HH:mm:ss
func (m *Model) UpdateTime() string {
	if m.UpdatedAt == 0 {
		return ""
	}
	return time.UnixMilli(m.UpdatedAt).Format(time.DateTime)
}
