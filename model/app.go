package model

type AppOtp struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(50);primary_key;"`
	GormModel

	AppKey      string `json:"app_key" query:"app_key" gorm:"type:varchar(50)"`
	AppSecret   string `json:"app_secret" query:"app_secret" gorm:"type:text"`
	Name        string `json:"name" query:"name" gorm:"type:varchar(200)"`
	Description string `json:"description" query:"description" gorm:"type:varchar(100)"`
}

type AppNotification struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(50);primary_key;"`
	GormModel

	AppKey      string `json:"app_key" query:"app_key" gorm:"type:varchar(50)"`
	AppSecret   string `json:"app_secret" query:"app_secret" gorm:"type:text"`
	Name        string `json:"name" query:"name" gorm:"type:varchar(200)"`
	Description string `json:"description" query:"description" gorm:"type:varchar(100)"`
}
