package orm

import "github.com/FourWD/middleware/model"

type LogAction struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	UserID    string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	Remark    string `json:"remark" query:"remark" gorm:"type:varchar(100)"`
	RemarkKey string `json:"remark_key" query:"remark_key" gorm:"type:varchar(100)"`
}
