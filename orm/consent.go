package orm

import "github.com/FourWD/middleware/model"

type Consent struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ConsentTypeID string `json:"consent_type_id" query:"consent_type_id" gorm:"type:varchar(36)"`

	ConsentVersion string `json:"consent_version" query:"consent_version" gorm:"type:varchar(3)"`
	Detail         string `json:"detail" query:"detail" gorm:"type:text"`
	DetailEn       string `json:"detail_en" query:"detail_en" gorm:"type:text"`
}
