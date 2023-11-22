package orm

import "github.com/FourWD/middleware/model"

type TermsCondition struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	TCVersion   string `json:"tc_version" query:"tc_version" gorm:"type:varchar(20)"`
	Description string `json:"description" query:"description" gorm:"type:text"`
}
