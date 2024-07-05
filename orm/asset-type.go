package orm

import "github.com/FourWD/middleware/model"

type AssetType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel

	Code        string `json:"code" query:"code" gorm:"type:varchar(150)"`
	Name        string `json:"name" query:"name" gorm:"type:varchar(150)"`
	Description string `json:"description" query:"description" gorm:"type:varchar(150)"`

	model.GormRowOrder
}
