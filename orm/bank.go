package orm

import (
	"github.com/FourWD/middleware/model"
)

type Bank struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name   string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`
	Logo   string `json:"logo" query:"logo" gorm:"type:varchar(256)"`

	model.GormRowOrder
}
