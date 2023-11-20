package orm

import (
	"time"
)

type LogEmployeeLoing struct {
	EmployeeID string    `json:"employee_id" query:"employee_id" gorm:"type:varchar(36);primary_key;"`
	CreatedAt  time.Time `json:"created_at" query:"created_at" gorm:"<-:create"`

	Status     string    `json:"status" query:"status" gorm:"type:varchar(20)"`
	TimeStamp  time.Time `json:"time_stamp" query:"time_stamp"`
	ExpireTime time.Time `json:"expire_time" query:"expire_time"`
}
