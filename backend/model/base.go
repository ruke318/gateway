package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// 时间格式常量
const TimeFormat = "2006-01-02 15:04:05"

// LocalTime 自定义时间类型，JSON序列化为 yyyy-mm-dd HH:mm:ss 格式
type LocalTime time.Time

// MarshalJSON 序列化为 JSON
func (t LocalTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, tt.Format(TimeFormat))), nil
}

// UnmarshalJSON 从 JSON 反序列化
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	// 去掉引号
	str := string(data[1 : len(data)-1])
	tt, err := time.ParseInLocation(TimeFormat, str, time.Local)
	if err != nil {
		return err
	}
	*t = LocalTime(tt)
	return nil
}

// Value 实现 driver.Valuer 接口，写入数据库
func (t LocalTime) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return nil, nil
	}
	return tt, nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取
func (t *LocalTime) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		*t = LocalTime(val)
	case []byte:
		tt, err := time.ParseInLocation(TimeFormat, string(val), time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(tt)
	case string:
		tt, err := time.ParseInLocation(TimeFormat, val, time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(tt)
	default:
		return fmt.Errorf("cannot scan type %T into LocalTime", v)
	}
	return nil
}

// BaseModel 公共字段
type BaseModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Status    int8      `gorm:"default:1;comment:状态：1启用 0禁用" json:"status"`
	CreatedAt LocalTime `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt LocalTime `gorm:"autoUpdateTime" json:"updated_at"`
}

// 状态常量
const (
	StatusEnabled  int8 = 1
	StatusDisabled int8 = 0
)
