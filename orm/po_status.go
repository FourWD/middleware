package orm

import "github.com/FourWD/middleware/model"

// midOrm "github.com/FourWD/middleware/orm"

type PoStatus struct { //
	ID string `json:"id" query:"id" gorm:"type:varchar(2); primary_key"`
	model.GormModel

	Name string `json:"name" query:"name" gorm:"type:varchar(256)"`
	model.GormRowOrder
}
