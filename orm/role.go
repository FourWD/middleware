package orm

import "github.com/FourWD/middleware/model"

type Role struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel
	Name string `json:"name" query:"name" gorm:"type:varchar(100)"`
	model.GormRowOrder
}
