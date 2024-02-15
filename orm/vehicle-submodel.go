package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleSubModel struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel
	VehicleModelID string `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36)"`
	Name           string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn         string `json:"name_en" query:"name_en" gorm:"type:varchar(20)"`

	model.GormRowOrder
}
