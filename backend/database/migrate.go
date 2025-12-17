package database

import (
	"github.com/ruke318/gateway/model"
)

// AutoMigrate 自动迁移表结构
func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},             // 用户表
		&model.OperationLog{},     // 操作日志表
		&model.Vendor{},           // 厂商表
		&model.Organization{},     // 机构表
		&model.ScriptLibrary{},    // 公共函数库表
		&model.HookScript{},       // Hook 脚本表
		&model.Service{},          // 接口配置表
		&model.ServiceHook{},      // 接口 Hook 关联表
		&model.DictionaryConfig{}, // 字典配置表
	)
}
