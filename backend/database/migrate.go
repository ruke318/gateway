package database

import (
	"github.com/ruke318/gateway/model"
)

// AutoMigrate 自动迁移表结构
func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.Vendor{},
		&model.Organization{},
		&model.ScriptLibrary{},
		&model.HookScript{},
		&model.Service{},
		&model.ServiceHook{},
	)
}
