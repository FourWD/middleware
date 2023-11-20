package orm

import (
	"time"
)

type Role struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`

	CreatedAt time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`

	Name     string `json:"file_name" query:"file_name" gorm:"type:varchar(100)"`
	RowOrder int    `json:"row_order" query:"row_order" gorm:"type:tinyint;"`
}
