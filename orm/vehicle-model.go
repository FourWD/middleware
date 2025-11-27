package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type VehicleModel struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	VehicleTypeID  string `json:"vehicle_type_id" query:"vehicle_type_id" gorm:"type:varchar(10)"`
	VehicleBrandID string `json:"vehicle_brand_id" query:"vehicle_brand_id" gorm:"type:varchar(36)"`
	Name           string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn         string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`

	OptionalID1 string `json:"optional_id_1" query:"optional_id_1" gorm:"column:optional_id_1;type:varchar(20)"`
	OptionalID2 string `json:"optional_id_2" query:"optional_id_2" gorm:"column:optional_id_2;type:varchar(20)"`
	OptionalID3 string `json:"optional_id_3" query:"optional_id_3" gorm:"column:optional_id_3;type:varchar(20)"`

	Remark string `json:"remark" query:"remark" gorm:"type:text"`

	SyncDate time.Time `json:"sync_date" query:"sync_date"`

	model.GormRowOrder
}
