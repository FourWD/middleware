package orm

import "github.com/FourWD/middleware/model"

type RolePermission struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	RoleID     string `json:"role_id" query:"role_id" gorm:"type:varchar(36); uniqueIndex:idx_role_permissions"`
	RoleMenuID string `json:"role_menu_id" query:"role_menu_id" gorm:"type:varchar(36); uniqueIndex:idx_role_permissions"`
	IsCreate   bool   `json:"is_create" query:"is_create" gorm:"type:bool;"`
	IsRead     bool   `json:"is_read" query:"is_read" gorm:"type:bool;"`
	IsUpdate   bool   `json:"is_update" query:"is_update" gorm:"type:bool;"`
	IsDelete   bool   `json:"is_delete" query:"is_delete" gorm:"type:bool;"`
}
