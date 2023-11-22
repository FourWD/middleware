package orm

import (
	"github.com/FourWD/middleware/model"
)

type LogUserLogin struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	UserID            string  `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	DeviceID          string  `json:"device_id" query:"device_id" gorm:"type:varchar(255)"`
	DeviceName        string  `json:"device_name" query:"device_name" gorm:"type:varchar(255)"`
	LocationName      string  `json:"location_name" query:"location_name" gorm:"type:varchar(255)"`
	Latitude          float64 `json:"latitude" query:"latitude" gorm:"type:float"`
	Longitude         float64 `json:"longitude" query:"longitude" gorm:"type:float"`
	Token             string  `json:"token" query:"token" gorm:"type:text"`
	NotificationToken string  `json:"notification_token" query:"notification_token" gorm:"type:varchar(255)"`
	IsActive          bool    `json:"is_active" query:"is_active" gorm:"type:bool"`
	IsLoginSuccess    bool    `json:"is_login_success" query:"is_login_success" gorm:"type:bool"`
	model.GormRowOrder
}
