package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type LogOTP struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	EmployeeID  string    `json:"employee_id" query:"employee_id" gorm:"type:varchar(36)"`
	UserID      string    `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
	DeviceID    string    `json:"device_id" query:"device_id" gorm:"type:varchar(255)"`
	OTP         string    `json:"otp" query:"otp" gorm:"type:varchar(6)"`
	RefCodeOTP  string    `json:"ref_code_otp" query:"ref_code_otp" gorm:"type:varchar(6)"`
	RequestDate time.Time `json:"request_date" query:"request_date"`
	SentDate    time.Time `json:"sent_date" query:"sent_date"`
	VerifyDate  time.Time `json:"verify_date" query:"verify_date"`
	ExpireDate  time.Time `json:"expire_date" query:"expire_date"`
}
