package orm

import "github.com/FourWD/middleware/model"

type FinanceInterest struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36)"`
	model.GormModel
	FinanceGroupID string  `json:"finance_group_id" query:"finance_group_id" gorm:"type:varchar(36)"`
	IMonth         int     `json:"i_month" query:"i_month" gorm:"type:int"`
	Interest       float64 `json:"interest" query:"interest" gorm:"type:decimal(5,3)"`
	DownPercent    float64 `json:"down_percent" query:"down_percent" gorm:"type:decimal(14,2)"`
}
