package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleBrand struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Name   string `json:"name" query:"name" gorm:"type:varchar(20)"`
	NameEn string `json:"name_en" query:"name_en" gorm:"type:varchar(20)"`

	OptionalID1 string `json:"optional_id_1" query:"optional_id_1" gorm:"column:optional_id_1;type:varchar(20)"`
	OptionalID2 string `json:"optional_id_2" query:"optional_id_2" gorm:"column:optional_id_2;type:varchar(20)"`
	OptionalID3 string `json:"optional_id_3" query:"optional_id_3" gorm:"column:optional_id_3;type:varchar(20)"`

	model.GormRowOrder
}
