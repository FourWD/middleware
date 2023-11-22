package orm

import "github.com/FourWD/middleware/model"

type LogUserStatus struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	UserID          string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	OldUserStatusID string `json:"old_user_status_id" query:"old_user_status_id" gorm:"type:varchar(2)"`
	NewUserStatusID string `json:"new_user_status_id" query:"new_user_status_id" gorm:"type:varchar(2)"`
}
