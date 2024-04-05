package orm

import "github.com/FourWD/middleware/model"

type RoleTemplate struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	RoleModuleID string `json:"role_module_id" query:"role_module_id" gorm:"type:varchar(36);"`
	Name         string `json:"name" query:"name" gorm:"type:varchar(100);"`
	IsActive     bool   `json:"is_active" query:"is_active" gorm:"type:bool;"`
	model.GormRowOrder
}
