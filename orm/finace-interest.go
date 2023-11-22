package orm

import "github.com/FourWD/middleware/model"

type FinaceInterest struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`
	model.GormModel
	FinaceID   string  `json:"finace_id" query:"finace_id" gorm:"type:varchar(36)"`
	IMonth     int     `json:"i_month" query:"i_month" gorm:"type:int"`
	Interest   float64 `json:"interest" query:"interest" gorm:"type:decimal(5,3)"`
	DownAmount float64 `json:"down_amount" query:"down_amount" gorm:"type:decimal(14,2)"`
}
