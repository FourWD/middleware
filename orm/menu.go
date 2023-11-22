package orm

import (
	"github.com/FourWD/middleware/model"
)

type Menu struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	MenuGroupID string `json:"menu_grounp_id" query:"menu_grounp_id" gorm:"type:varchar(36)"`

	Subject      string `json:"name" query:"name" gorm:"type:varchar(500)"`
	Description  string `json:"description" query:"description" gorm:"type:varchar(500)"`
	ImagePath    string `json:"image_path" query:"image_path" gorm:"type:varchar(255)"`
	Url          string `json:"url" query:"url" gorm:"type:varchar(200)"`
	YoutubeUrl   string `json:"youtube_url" query:"youtube_url" gorm:"type:varchar(100)"`
	OpenType     string `json:"open_type" query:"open_type" gorm:"type:varchar(11)"`
	IsAppUrl     bool   `json:"is_app_url" query:"is_app_url" gorm:"type:bool"`

	model.GormRowOrder
}
