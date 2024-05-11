package orm

import "github.com/FourWD/middleware/model"

type District struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(4);primary_key"`
	model.GormModel

	ProvinceID string `json:"province_id" query:"province_id" gorm:"type:varchar(2)"`
	Name       string `json:"name" query:"name" gorm:"not null;type:varchar(100)"`
	NameEn     string `json:"name_en" query:"name_en" gorm:"type:varchar(100)"`

	model.GormRowOrder
}
