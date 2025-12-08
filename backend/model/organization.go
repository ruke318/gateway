package model

import "encoding/json"

// Organization 机构
type Organization struct {
	BaseModel
	Code        string          `gorm:"size:64;uniqueIndex;not null;comment:机构编码" json:"code"`
	Name        string          `gorm:"size:128;not null;comment:机构名称" json:"name"`
	Config      json.RawMessage `gorm:"type:json;comment:机构配置" json:"config"`
	Description string          `gorm:"type:text;comment:描述" json:"description"`
}

func (Organization) TableName() string {
	return "organization"
}
