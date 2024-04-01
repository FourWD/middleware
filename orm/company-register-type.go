package orm

import (
	"github.com/FourWD/middleware/model"
)

type CompanyRegisterType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Name       string `json:"name" query:"name" gorm:"type:varchar(100)"`
	UserTypeID string `json:"user_type_id" query:"user_type_id" gorm:"type:varchar(10)"`
	model.GormRowOrder
}
