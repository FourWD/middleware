package orm

type RolePermission struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	GormModel
	RoleID         string `json:"role_id" query:"role_id" gorm:"type:varchar(36); uniqueIndex:idx_role_permissions"`
	RoleTemplateID string `json:"role_template_id" query:"role_template_id" gorm:"type:varchar(36); uniqueIndex:idx_role_permissions"`
	IsCreate       bool   `json:"is_create" query:"is_create" gorm:"type:bool;"`
	IsRead         bool   `json:"is_read" query:"is_read" gorm:"type:bool;"`
	IsUpdate       bool   `json:"is_update" query:"is_update" gorm:"type:bool;"`
	IsDelete       bool   `json:"is_delete" query:"is_delete" gorm:"type:bool;"`
}
