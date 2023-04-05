package orm

import (
	"time"

	"gorm.io/gorm"
)

type GormModel struct {
	CreatedAt time.Time      `db:"created_at" json:"created_at" gorm:"<-:create"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at" json:"deleted_at" gorm:"index"`

	CreatedBy string `db:"created_by"  json:"created_by" gorm:"type:varchar(36)"`
	UpdatedBy string `db:"updated_by"  json:"updated_by" gorm:"type:varchar(36)"`
	DeletedBy string `db:"deleted_by"  json:"deleted_by" gorm:"type:varchar(36)"`
}
