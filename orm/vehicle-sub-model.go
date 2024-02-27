package orm

import (
	"github.com/FourWD/middleware/model"
)

type VehicleSubModel struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel
	VehicleModelID string `json:"vehicle_model_id" query:"vehicle_model_id" gorm:"type:varchar(36)"`
	Name           string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn         string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`
	Image1         string `json:"image_1" query:"image_1" gorm:"column:image_1;type:varchar(255)"`
	Image2         string `json:"image_2" query:"image_2" gorm:"column:image_2;type:varchar(255)"`
	Image3         string `json:"image_3" query:"image_3" gorm:"column:image_3;type:varchar(255)"`

	model.GormRowOrder
}