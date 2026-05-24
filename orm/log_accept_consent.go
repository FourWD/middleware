package orm

import "github.com/FourWD/middleware/model"

type LogAcceptConsent struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ConsentID string `json:"consent_id" query:"consent_id" gorm:"type:varchar(36)"`
	UserID    string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
}
