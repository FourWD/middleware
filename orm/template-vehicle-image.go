package orm

import (
	"github.com/FourWD/middleware/model"
)

type TemplateVehicleImage struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	VehicleImageGroupID string `json:"vehicle_type_id" query:"vehicle_type_id" gorm:"type:varchar(36)"`
	Name                string `json:"name" query:"name" gorm:"type:varchar(255)"`
	model.GormRowOrder
}

// รูป ด้านซ้าย ด้านขวา ของภายหน้า
