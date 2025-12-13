package model

import (
	"golang.org/x/crypto/bcrypt"
)

// User 用户表
type User struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string     `gorm:"size:64;uniqueIndex;not null;comment:用户名" json:"username"`
	Password  string     `gorm:"size:128;not null;comment:密码（bcrypt加密）" json:"-"` // json:"-" 不返回密码
	RealName  string     `gorm:"size:128;comment:真实姓名" json:"real_name"`
	Role      string     `gorm:"size:32;default:user;comment:角色:admin管理员/user普通用户" json:"role"`
	Status    int        `gorm:"default:1;comment:状态:1启用/0禁用" json:"status"`
	LastLogin *LocalTime `gorm:"comment:最后登录时间" json:"last_login"` // 使用指针类型，允许 NULL
	CreatedAt LocalTime  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt LocalTime  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}

// SetPassword 设置密码（bcrypt 加密）
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin 是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// OperationLog 操作日志表
type OperationLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;comment:操作用户ID" json:"user_id"`
	Username   string    `gorm:"size:64;comment:操作用户名" json:"username"`
	Operation  string    `gorm:"size:32;comment:操作类型:create/update/delete" json:"operation"`
	Resource   string    `gorm:"size:64;comment:资源类型:vendor/organization/service等" json:"resource"`
	ResourceID string    `gorm:"size:128;comment:资源ID" json:"resource_id"`
	BeforeData string    `gorm:"type:text;comment:修改前数据JSON（update/delete操作）" json:"before_data"`
	AfterData  string    `gorm:"type:text;comment:修改后数据JSON（create/update操作）" json:"after_data"`
	IP         string    `gorm:"size:128;comment:操作IP" json:"ip"`
	CreatedAt  LocalTime `gorm:"autoCreateTime;index" json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_log"
}
