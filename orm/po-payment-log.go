package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

// midOrm "github.com/FourWD/middleware/orm"

type PoPaymentLog struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36); uniqueIndex:idx_id "` //
	model.GormModel

	Title         string    `json:"title" query:"title" gorm:"type:varchar(50)"`
	Price         float64   `json:"price" query:"price" gorm:"type:decimal(14,2)"`
	PaymentTypeID string    `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(2)"`
	PaymentDate   time.Time `json:"payment_date" query:"payment_date"`
}
