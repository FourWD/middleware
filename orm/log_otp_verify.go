package orm

import (
	"time"
)

type LogOtpVerify struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key;"`

	CreatedAt time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`

	AppID     string `json:"app_id" query:"app_id" gorm:"type:text"`
	AppKey    string `json:"app_key" query:"app_key" gorm:"type:varchar(50)"`
	AppSecret string `json:"app_secret" query:"app_secret" gorm:"type:text"`
	Payload   string `json:"payload" query:"payload" gorm:"type:text"`
	Response  string `json:"response" query:"response" gorm:"type:text"`
}
