package orm

import (
	"time"

	"github.com/FourWD/middleware/model"
)

type WelcomeMessage struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	StartDate       time.Time `json:"start_date" query:"start_date"`
	EndDate         time.Time `json:"end_date" query:"end_date"`
	IsShow          bool      `json:"is_show" query:"is_show" gorm:"bool"`
	ArticleID       string    `json:"article_id" query:"article_id" gorm:"type:varchar(36)"`
	CustomImagePath string    `json:"custom_image_path" query:"custom_image_path" gorm:"type:varchar(50)"`
}
