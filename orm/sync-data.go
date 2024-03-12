package orm

import (
	"time"
)

type SyncData struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`

	Name       string    `json:"name" query:"name" gorm:"not null;type:varchar(50)"`
	StartDate  time.Time `json:"start_date" query:"start_date"`
	EndDate    time.Time `json:"end_date" query:"end_date"`
	IsManual   bool      `json:"is_manual" query:"is_manual" gorm:"type:bool"`
	IsComplete bool      `json:"is_complete" query:"is_complete" gorm:"type:bool"`
}
