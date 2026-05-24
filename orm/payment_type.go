package orm

import (
	"github.com/FourWD/middleware/model"
)

type PaymentType struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(2);primary_key"`
	model.GormModel

	Code   string `json:"code" query:"code" gorm:"type:varchar(2)"`
	Logo   string `json:"logo" query:"logo" gorm:"type:varchar(255)"`
	Name   string `json:"name" query:"name" gorm:"type:varchar(50)"`
	NameEn string `json:"name_en" query:"name_en" gorm:"type:varchar(50)"`

	PaymentGroupID string `json:"payment_group_id" query:"payment_group_id" gorm:"type:varchar(10)"`
	IsActive       bool   `json:"is_active" query:"is_active" gorm:"type:bool"`
	IsDeposit      bool   `json:"is_deposit" query:"is_deposit" gorm:"type:bool"`

	model.GormRowOrder
}

/*
01 เงินสด PAYMENT
02 ขอสินเชื่อ PAYMENT
03 โอนเงิน REGISTER
04 โมบายแบ๊งค์กิ้งค์ REGISTER
*/
