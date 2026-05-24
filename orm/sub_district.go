package orm

import (
	"github.com/FourWD/middleware/model"
)

type SubDistrict struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(6);primary_key"`
	model.GormModel

	DistrictID string `json:"district_id" query:"district_id" gorm:"type:varchar(4)"`
	Name       string `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
	NameEn     string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`
	Postcode   string `json:"postcode" query:"postcode" gorm:"type:varchar(5)"`
	model.GormRowOrder
}
