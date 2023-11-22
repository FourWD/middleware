package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleBrand struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Name string `json:"name" query:"name" gorm:"type:varchar(20)"`
	model.GormRowOrder
}
