package orm

import "github.com/FourWD/middleware/model"

type MenuGroup struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	Code string `json:"code" query:"code" gorm:"unique;index;type:varchar(45)"`
	Name string `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
}
