package orm

import "github.com/FourWD/middleware/model"

type Role struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	Name         string `json:"name" query:"name" gorm:"type:varchar(100)"`
	RoleModuleID string `json:"role_module_id" query:"role_module_id" gorm:"type:varchar(36);"`
	model.GormRowOrder
}
