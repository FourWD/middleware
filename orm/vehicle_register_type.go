package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleRegisterType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel
	Code string `json:"code" query:"code" gorm:"type:varchar(10)"`
	Name string `json:"name" query:"name" gorm:"type:varchar(50)"`
	model.GormRowOrder
}
