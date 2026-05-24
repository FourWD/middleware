package orm

import "github.com/FourWD/middleware/model"

type LogSurveyAnswer struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	SurveyGroupID string `json:"survey_group_id" query:"survey_group_id" gorm:"type:varchar(36)"`
	UserID        string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
}
