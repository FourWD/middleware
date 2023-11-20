package orm

import (
	"time"
)

type RoleTemplate struct {
	ID   string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	Name string `json:"name" query:"name" gorm:"type:varchar(100);"`

	CreatedAt time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`
}
