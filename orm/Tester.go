package orm

import (
	"github.com/FourWD/middleware/model"
)

type Tester struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel
	Firstname string `json:"firstname" query:"firstname" gorm:"type:varchar(100)"`
	Lastname  string `json:"lastname" query:"lastname" gorm:"type:varchar(100)"`
	Mobile    string `json:"mobile" query:"mobile" gorm:"type:varchar(20);unique"`
}
