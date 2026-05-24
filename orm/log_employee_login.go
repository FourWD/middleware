package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type LogEmployeeLogin struct {
	EmployeeID string `json:"employee_id" query:"employee_id" gorm:"type:varchar(36);primary_key;"`
	model.GormModel

	Status     string    `json:"status" query:"status" gorm:"type:varchar(20)"`
	TimeStamp  time.Time `json:"time_stamp" query:"time_stamp"`
	ExpireTime time.Time `json:"expire_time" query:"expire_time"`
}
