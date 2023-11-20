package orm

import (
	"time"
)

type RolePermission struct {
	RoleID     string    `json:"role_id" query:"role_id" gorm:"type:varchar(36);primary_key;"`
	TemplateID string    `json:"template_id" query:"template_id" gorm:"type:varchar(100);"`
	IsCreate   bool      `json:"is_create" query:"is_create" gorm:"type:bool;"`
	IsRead     bool      `json:"is_read" query:"is_read" gorm:"type:bool;"`
	IsUpdate   bool      `json:"is_update" query:"is_update" gorm:"type:bool;"`
	IsDelete   bool      `json:"is_delete" query:"is_delete" gorm:"type:bool;"`
	CreatedAt  time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`
}
