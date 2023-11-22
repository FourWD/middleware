package orm

import (
	"github.com/FourWD/middleware/model"
)

type Article struct {
	ID string `json:"id" query:"id" gorm:"type:varchar(36);primary_key"`
	model.GormModel

	ArticleTypeID string `json:"article_type_id" query:"article_type_id" gorm:"type:varchar(36)"`
	FileCoverID   string `json:"file_cover_id" query:"file_cover_id" gorm:"type:varchar(36)"`
	Subject       string `json:"subject" query:"subject" gorm:"type:varchar(500)"`
	Detail        string `json:"detail" query:"detail" gorm:"type:text"`
	ButtonName    string `json:"button_name" query:"button_name" gorm:"type:varchar(100)"`
	ButtonUrl     string `json:"button_url" query:"button_url" gorm:"type:varchar(100)"`
	IsShowOnApp   bool   `json:"is_show_on_app" query:"is_show_on_app" gorm:"type:bool"`
	Tag           string `json:"tag" query:"tag" gorm:"type:varchar(100)"`
	CountView     int    `json:"count_view" query:"count_view" gorm:"type:int"`
	model.GormRowOrder
}
