package orm

import (
	"github.com/FourWD/middleware/model"
)

type OpenType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(11);primary_key"`
	model.GormModel

	Name string `json:"name" query:"name" gorm:"type:varchar(20)"`
	model.GormRowOrder
}
