package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

// midOrm "github.com/FourWD/middleware/orm"

type PoPayment struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`
	model.GormModel
	PaymentTypeID string    `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(2)"`
	PaymentDate   time.Time `json:"payment_date" query:"payment_date"`
	Amount        float64   `json:"amount" query:"amount" gorm:"type:decimal(14,2)"`
	Remark        string    `json:"remark" query:"remark" gorm:"type:varchar(50)"`
}
