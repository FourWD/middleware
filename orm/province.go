package orm

import (
	"github.com/FourWD/middleware/model"
)

type Province struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name      string `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
	NameShort string `json:"name_short" query:"name_short" gorm:"type:varchar(50)"`
	NameEn    string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`
	model.GormRowOrder
}
