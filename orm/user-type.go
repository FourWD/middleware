package orm

import (
	"github.com/FourWD/middleware/model"
)

type UserType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Code   string `json:"code" query:"code" gorm:"type:varchar(2)"`
	Name   string `json:"name" query:"name" gorm:"type:varchar(20)"`
	NameEn string `json:"name_en" query:"name_en" gorm:"type:varchar(20)"`

	model.GormRowOrder
}
