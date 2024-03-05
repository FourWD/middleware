package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Maintenance struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	MaintenanceType       string    `json:"maintenance_type" query:"maintenance_type" gorm:"type:varchar(500)"`
	MaintenanceLocationID string    `json:"maintenance_location_id" query:"maintenance_location_id" gorm:"type:varchar(36)"`
	ChassisNumber         string    `json:"chassis_number" query:"chassis_number" gorm:"type:varchar(20)"`
	Mile                  int       `json:"mile" query:"mile" gorm:"type:int"`
	MaintenanceDate       time.Time `json:"maintenance_date" query:"maintenance_date"`
	Detail                string    `json:"detail" query:"detail" gorm:"type:varchar(500)"`

	model.GormRowOrder
}
