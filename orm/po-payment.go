package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

// midOrm "github.com/FourWD/middleware/orm"

type PoPayment struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`

	model.GormModel

	PoID          string    `json:"po_id" query:"po_id" gorm:"type:varchar(36)"`
	PoNo          string    `json:"po_no" query:"po_no" gorm:"type:varchar(20)"`
	PaymentTypeID string    `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(2)"`
	PaymentDate   time.Time `json:"payment_date" query:"payment_date"`
	Amount        float64   `json:"amount" query:"amount" gorm:"type:decimal(14,2)"`
	RemainAmount  float64   `json:"remain_amount" query:"remain_amount" gorm:"type:decimal(14,2)"`
	ImageUrl      string    `json:"image_url" query:"image_url" gorm:"type:varchar(1000)"`
	Remark        string    `json:"remark" query:"remark" gorm:"type:varchar(50)"`
}
