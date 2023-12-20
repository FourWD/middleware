package orm

import "github.com/FourWD/middleware/model"

type Callback struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	CallbackTypeID  string `json:"callback_type_id" query:"callback_type_id" gorm:"type:varchar(2)"`
	RequestDate     string `json:"requset_date" query:"requset_date" gorm:"type:text"`
	RequestUserID   string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	RequestMessage  string `json:"requset_message" query:"requset_message" gorm:"type:text"`
	IsReponse       bool   `json:"is_reponse" query:"is_reponse" gorm:"type:bool"`
	ResponseDate    string `json:"response_date" query:"response_date" gorm:"type:text"`
	ResponseUserID  string `json:"response_user_id" query:"response_user_id" gorm:"type:varchar(36)"`
	ResponseMessage string `json:"response_message" query:"response_message" gorm:"type:text"`
}
