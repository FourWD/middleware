package orm

import "github.com/FourWD/middleware/model"

type PaymentSubType struct { //เก็บ visa car , master car etc.
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	PaymentTypeID string `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(10)"`
	Code          string `json:"code" query:"code" gorm:"type:varchar(2)"`
	Logo          string `json:"logo" query:"logo" gorm:"type:varchar(255)"`
	Name          string `json:"name" query:"name" gorm:"type:varchar(50)"`
}
