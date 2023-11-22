package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleSubModelColor struct {
	VehicleSubModelID string `json:"vehicle_submodel_id" query:"vehicle_submodel_id" gorm:"type:varchar(2); uniqueIndex:idx_vehicle_submodel_color"`
	VehicleSubColor   string `json:"vehicle_color_id" query:"vehicle_color_id" gorm:"type:varchar(36);  uniqueIndex:idx_vehicle_submodel_color"`
	model.GormRowOrder
}
