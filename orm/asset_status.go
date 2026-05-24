package orm

import "github.com/FourWD/middleware/model"

type AssetStatus struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel

	Name string `json:"name" query:"name" gorm:"type:varchar(150)"`
	Icon string `json:"icon" query:"icon" gorm:"type:varchar(150)"`

	model.GormRowOrder
}
