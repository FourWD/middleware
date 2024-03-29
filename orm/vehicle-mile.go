package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleMile struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name      string `json:"name" query:"name" gorm:"type:varchar(20)"`
	StartMile string `json:"start_mile" query:"start_mile" gorm:"type:varchar(10)"`
	EndMile   string `json:"end_mile" query:"end_mile" gorm:"type:varchar(10)"`
	MileMin   int    `json:"mile_min" query:"mile_min" gorm:"type:int"`
	MileMax   int    `json:"mile_max" query:"mile_max" gorm:"type:int"`
	model.GormRowOrder
}
