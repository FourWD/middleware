package orm

import (
	"github.com/FourWD/middleware/model"
)

type SurveyOption struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	SurveyID  string `json:"survey_id" query:"survey_id" gorm:"type:varchar(36)"`
	Answer    string `json:"answer" query:"answer" gorm:"not null;type:varchar(100)"`
	ImagePath string `json:"image_path" query:"image_path" gorm:"type:varchar(255)"`
	model.GormRowOrder
}

// ชื่อของ survey เช่น survey "ตัวเลือก suv sedan"
