package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type VehicleMaintenance struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	VehicleID       string    `json:"vehicle_id" query:"vehicle_id" gorm:"type:varchar(36)"`
	MaintenanceDate time.Time `json:"maintenance_date" query:"maintenance_date"`
	MaintenanceLocationID string `json:"maintenance_location_id" query:"maintenance_location_id" gorm:"type:varchar(36)"`
	MaintenanceTypeID     string `json:"maintenance_type_id" query:"maintenance_type_id" gorm:"type:varchar(36)"`
	Mile                  int    `json:"mile" query:"mile" gorm:"type:int"`
	Remark                string `json:"remark" query:"remark" gorm:"type:text"`
}
