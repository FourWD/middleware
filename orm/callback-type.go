package orm

import (
	"github.com/FourWD/middleware/model"
)

type CallbackType struct { //ประเภทของการติดต่อ
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name string `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
	model.GormRowOrder
}
