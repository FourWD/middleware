package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type Notification struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ToUserID              string `json:"to_user_id" query:"to_user_id" gorm:"type:varchar(36)"`
	ToNotificationTopicID string `json:"to_notification_topic_id" query:"to_notification_topic_id" gorm:"type:varchar(36)"`

	NotificationTypeID string `json:"notification_type_id" query:"notification_type_id" gorm:"type:varchar(2)"`
	Message            string `json:"message" query:"message" gorm:"type:varchar(500)"`
	Title              string `json:"title" query:"title" gorm:"type:varchar(500)"`

	Url      string    `json:"url" query:"url" gorm:"type:varchar(500)"`
	ShowDate time.Time `json:"show_date" query:"show_date" firestore:"show_date"`
	// IsRead             bool      `json:"is_read" query:"is_read" gorm:"bool"`
	// ReadDate           time.Time `json:"read_date" query:"read_date"`
}
