package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleModel struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	VehicleTypeID  string `json:"vehicle_type_id" query:"vehicle_type_id" gorm:"type:varchar(10)"`
	VehicleBrandID string `json:"vehicle_brand_id" query:"vehicle_brand_id" gorm:"type:varchar(36)"`
	Name           string `json:"name" query:"name" gorm:"type:varchar(50)"`
	model.GormRowOrder
}
