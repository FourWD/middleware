package model

import (
	"time"

	"gorm.io/gorm"
)

type GormModel struct {
	CreatedAt time.Time      `json:"created_at" query:"created_at" gorm:"<-:create"`
	UpdatedAt time.Time      `json:"updated_at" query:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" query:"deleted_at" gorm:"index"`

	CreatedBy string `json:"created_by" query:"created_by" gorm:"type:varchar(36)"`
	UpdatedBy string `json:"updated_by" query:"updated_by" gorm:"type:varchar(36)"`
	DeletedBy string `json:"deleted_by" query:"deleted_by" gorm:"type:varchar(36)"`
}

type GormRowOrder struct {
	RowOrder int `json:"row_order" query:"row_order" gorm:"type:int(4);"`
}
