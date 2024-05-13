package orm

import (
	"github.com/FourWD/middleware/model"
)

type NotificationTopicUser struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	NotificationTopicID string `json:"notification_topic_id" query:"notification_topic_id" gorm:"type:varchar(36)"`
	UserID              string `json:"user_id" query:"user_id" gorm:"type:varchar(36)"`
}
