package orm

import "github.com/FourWD/middleware/model"

// midOrm "github.com/FourWD/middleware/orm"

type PoGift struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36); primary_key"`
	model.GormModel

	PoID        string  `json:"po_id" query:"po_id" gorm:"type:varchar(36)"`
	GiftID      string  `json:"gift_id" query:"gift_id" gorm:"type:varchar(36)"`
	UnitPreVat  float64 `json:"unit_price_pre_vat" query:"unit_price_pre_vat" gorm:"type:decimal(14,2)"`
	UnitVat     float64 `json:"unit_vat" query:"unit_vat" gorm:"type:decimal(14,2)"`
	UnitPrice   float64 `json:"unit_price" query:"unit_price" gorm:"type:decimal(14,2)"`
	Qty         int     `json:"qty" query:"qty" gorm:"type:int"`
	PricePreVat float64 `json:"price_pre_vat" query:"price_pre_vat" gorm:"type:decimal(14,2)"`
	Vat         float64 `json:"vat" query:"vat" gorm:"type:decimal(14,2)"`
	Price       float64 `json:"price" query:"price" gorm:"type:decimal(14,2)"`
	model.GormRowOrder
}
