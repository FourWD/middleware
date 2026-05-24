package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

// midOrm "github.com/FourWD/middleware/orm"

type PoPayment struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`

	model.GormModel

	AuctionID     string    `json:"auction_id" query:"auction_id" gorm:"type:varchar(36)"`
	UserID        string    `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	PaymentTypeID string    `json:"payment_type_id" query:"payment_type_id" gorm:"type:varchar(2)"`
	PaymentDate   time.Time `json:"payment_date" query:"payment_date"`
	Amount        float64   `json:"amount" query:"amount" gorm:"type:decimal(14,2)"`
	ImageUrl      string    `json:"image_url" query:"image_url" gorm:"type:varchar(1000)"`
	Remark        string    `json:"remark" query:"remark" gorm:"type:varchar(50)"`
	IsPaid        bool      `json:"is_paid" query:"is_paid" gorm:"type:bool;default:false"`
}
