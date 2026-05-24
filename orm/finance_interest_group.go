package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type FinanceInterestGroup struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`
	model.GormModel
	FinanceID string    `json:"finance_id" query:"finance_id" gorm:"type:varchar(36)"`
	StartDate time.Time `json:"start_date" query:"start_date"`
	EndDate   time.Time `json:"end_date" query:"end_date"`
}
