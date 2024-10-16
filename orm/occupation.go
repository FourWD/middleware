package orm

import (
	"github.com/FourWD/middleware/model"
)

type Occupation struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel
	Name   string `json:"name" query:"name" gorm:"type:varchar(300)"`
	NameEn string `json:"name_en" query:"name_en" gorm:"type:varchar(300)"`

	OccupationID string `json:"occupation_id" query:"occupation_id" gorm:"type:varchar(36)"`
	model.GormRowOrder
}
